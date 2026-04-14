package firebase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QuerySections queries SectionDoc documents.
// All query parameters are expected to be normalized by the handler
// Some parameters are filtered in-memory due to some weird Firestore limitations
func (c *Firestore) QuerySections(ctx context.Context, q types.SectionQuery) ([]types.SectionDoc, bool, error) {
	if q.Term == "" {
		return []types.SectionDoc{}, false, nil
	}

	key := cacheKey("sections", "term", q.Term, "prefix", q.Prefix, "number", q.Number, "instructor", q.Instructor, "days", q.Days, "building", q.Building, "title", q.Title, "limit", strconv.Itoa(q.Limit), "offset", strconv.Itoa(q.Offset))

	if cached, found := c.Cache.Get(key); found {
		log.Printf("Cache hit: %s", key)
		r := cached.(cachedResult[types.SectionDoc])
		return r.Items, r.HasNext, nil
	}

	// building and title are in-memory filters: everything else goes to Firestore
	if q.Building != "" || q.Title != "" {
		allSections, err := c.fetchAllSections(ctx, q)
		if err != nil {
			return nil, false, err
		}

		filtered := allSections
		if q.Building != "" {
			filtered = filterByBuilding(filtered, q.Building)
		}
		if q.Title != "" {
			filtered = filterByTitle(filtered, q.Title)
		}
		sections, hasNext := paginate(filtered, q.Limit, q.Offset)

		c.Cache.Set(key, cachedResult[types.SectionDoc]{Items: sections, HasNext: hasNext}, c.TTL.Sections)

		return sections, hasNext, nil
	}

	// No in-memory filters needed: let Firestore handle pagination
	query := c.buildSectionQuery(q.Term, q.Prefix, q.Number)
	if q.Instructor != "" {
		query = query.Where("instructor_name_normalized", "==", q.Instructor)
	}
	if q.Days != "" {
		query = query.Where("days", "array-contains", q.Days)
	}

	sections, hasNext, err := c.collectSections(ctx, query, q.Limit, q.Offset, false)
	if err != nil {
		return nil, false, err
	}

	c.Cache.Set(key, cachedResult[types.SectionDoc]{Items: sections, HasNext: hasNext}, c.TTL.Sections)

	return sections, hasNext, nil
}

// buildSectionQuery returns a base Firestore query filtered by term and optionally prefix/number.
func (c *Firestore) buildSectionQuery(term, prefix, number string) firestore.Query {
	if prefix != "" && number != "" {
		return c.Collection("courses").Doc(prefix).Collection("numbers").Doc(number).Collection("sections").Where("term", "==", term)
	}
	query := c.CollectionGroup("sections").Where("term", "==", term)
	if prefix != "" {
		query = query.Where("prefix", "==", prefix)
	}
	return query
}

func (c *Firestore) collectSections(ctx context.Context, query firestore.Query, limit, offset int, skipPagination bool) ([]types.SectionDoc, bool, error) {
	if !skipPagination && limit > 0 {
		query = query.Offset(offset).Limit(limit + 1)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var sections []types.SectionDoc
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to iterate sections: %w", err)
		}
		var s types.SectionDoc
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		sections = append(sections, s)
	}

	if skipPagination {
		return sections, false, nil
	}

	hasNext := limit > 0 && len(sections) > limit
	if hasNext {
		sections = sections[:limit]
	}
	return sections, hasNext, nil
}

// fetchAllSections returns all sections for the Firestore-level filters, caching the unpaginated result so in-memory filters can work properly
// The cache key excludes building since that is filtered after fetching.
func (c *Firestore) fetchAllSections(ctx context.Context, q types.SectionQuery) ([]types.SectionDoc, error) {
	allKey := cacheKey("sections_all", "term", q.Term, "prefix", q.Prefix, "number", q.Number, "instructor", q.Instructor, "days", q.Days)

	if cached, found := c.Cache.Get(allKey); found {
		log.Printf("Cache hit: %s", allKey)
		return cached.([]types.SectionDoc), nil
	}

	query := c.buildSectionQuery(q.Term, q.Prefix, q.Number)
	if q.Instructor != "" {
		query = query.Where("instructor_name_normalized", "==", q.Instructor)
	}
	if q.Days != "" {
		query = query.Where("days", "array-contains", q.Days)
	}

	sections, _, err := c.collectSections(ctx, query, 0, 0, true)
	if err != nil {
		return nil, err
	}

	c.Cache.Set(allKey, sections, c.TTL.Sections)
	return sections, nil
}

func filterByTitle(sections []types.SectionDoc, title string) []types.SectionDoc {
	var out []types.SectionDoc
	for _, s := range sections {
		if strings.Contains(strings.ToLower(s.Title), title) {
			out = append(out, s)
		}
	}
	return out
}

func filterByBuilding(sections []types.SectionDoc, building string) []types.SectionDoc {
	var out []types.SectionDoc
	for _, s := range sections {
		if strings.EqualFold(s.Building, building) {
			out = append(out, s)
		}
	}
	return out
}

// pagination is different for in-memory filters
func paginate[T any](items []T, limit, offset int) ([]T, bool) {
	if limit <= 0 {
		return items, false
	}
	if offset >= len(items) {
		return nil, false
	}
	end := offset + limit
	hasNext := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], hasNext
}

// GetSectionByParams retrieves a single SectionDoc via a direct document path.
// The section_address document ID is constructed as "{prefix}{number}.{section}.{term}".
func (c *Firestore) GetSectionByParams(ctx context.Context, prefix, number, term, section string) (*types.SectionDoc, error) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	number = strings.ToLower(strings.TrimSpace(number))
	term = strings.ToLower(strings.TrimSpace(term))
	section = strings.ToLower(strings.TrimSpace(section))

	sectionAddress := prefix + number + "." + section + "." + term

	key := cacheKey("sections", "section_address", sectionAddress)
	if cached, found := c.Cache.Get(key); found {
		if s, ok := cached.(types.SectionDoc); ok {
			return &s, nil
		}
	}

	doc, err := c.Collection("courses").Doc(prefix).Collection("numbers").Doc(number).Collection("sections").Doc(sectionAddress).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	var s types.SectionDoc
	if err := doc.DataTo(&s); err != nil {
		return nil, fmt.Errorf("failed to parse section: %w", err)
	}

	c.Cache.Set(key, s, c.TTL.Sections)

	return &s, nil
}

// QueryGeneralCourses queries CourseGeneralInfo documents.
// If prefix is provided, the query is scoped to courses/{prefix}/numbers.
// Otherwise a collection group query across all "numbers" subcollections is used.
func (c *Firestore) QueryGeneralCourses(ctx context.Context, q types.GeneralCourseQuery) ([]types.CourseGeneralInfo, bool, error) {
	key := cacheKey("courses", "prefix", q.Prefix, "search", q.Search, "limit", strconv.Itoa(q.Limit), "offset", strconv.Itoa(q.Offset))

	if cached, found := c.Cache.Get(key); found {
		log.Printf("Cache hit: %s", key)
		r := cached.(cachedResult[types.CourseGeneralInfo])
		return r.Items, r.HasNext, nil
	}

	if q.Search != "" {
		allCourses, err := c.fetchAllCourses(ctx, q.Prefix)
		if err != nil {
			return nil, false, err
		}

		search := strings.ToLower(q.Search)
		var filtered []types.CourseGeneralInfo
		for _, course := range allCourses {
			if strings.Contains(strings.ToLower(course.Title), search) || strings.Contains(strings.ToLower(course.Description), search) {
				filtered = append(filtered, course)
			}
		}

		result, hasNext := paginate(filtered, q.Limit, q.Offset)
		if result == nil {
			result = []types.CourseGeneralInfo{}
		}
		c.Cache.Set(key, cachedResult[types.CourseGeneralInfo]{Items: result, HasNext: hasNext}, c.TTL.Courses)
		return result, hasNext, nil
	}

	var query firestore.Query
	if q.Prefix != "" {
		query = c.Collection("courses").Doc(q.Prefix).Collection("numbers").Query
	} else {
		query = c.CollectionGroup("numbers").Query
	}

	courses, hasNext, err := c.collectGeneralCourses(ctx, query, q.Limit, q.Offset, false)
	if err != nil {
		return nil, false, err
	}

	c.Cache.Set(key, cachedResult[types.CourseGeneralInfo]{Items: courses, HasNext: hasNext}, c.TTL.Courses)
	return courses, hasNext, nil
}

// fetchAllCourses returns all CourseGeneralInfo documents for a prefix, caching the
// unpaginated result so search queries can filter in memory without re-fetching Firestore.
func (c *Firestore) fetchAllCourses(ctx context.Context, prefix string) ([]types.CourseGeneralInfo, error) {
	allKey := cacheKey("courses_all", "prefix", prefix)
	if cached, found := c.Cache.Get(allKey); found {
		log.Printf("Cache hit: %s", allKey)
		return cached.([]types.CourseGeneralInfo), nil
	}

	var query firestore.Query
	if prefix != "" {
		query = c.Collection("courses").Doc(prefix).Collection("numbers").Query
	} else {
		query = c.CollectionGroup("numbers").Query
	}

	courses, _, err := c.collectGeneralCourses(ctx, query, 0, 0, true)
	if err != nil {
		return nil, err
	}

	c.Cache.Set(allKey, courses, c.TTL.Courses)
	return courses, nil
}

// GetGeneralCourse retrieves a single CourseGeneralInfo at courses/{prefix}/numbers/{number}.
func (c *Firestore) GetGeneralCourse(ctx context.Context, prefix, number string) (*types.CourseGeneralInfo, error) {
	if prefix == "" || number == "" {
		return nil, nil
	}

	key := cacheKey("courses", "prefix", prefix, "number", number)
	if cached, found := c.Cache.Get(key); found {
		if course, ok := cached.(types.CourseGeneralInfo); ok {
			return &course, nil
		}
	}

	doc, err := c.Collection("courses").Doc(prefix).Collection("numbers").Doc(number).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	var course types.CourseGeneralInfo
	if err := doc.DataTo(&course); err != nil {
		return nil, fmt.Errorf("failed to parse course: %w", err)
	}

	c.Cache.Set(key, course, c.TTL.Courses)

	return &course, nil
}

func (c *Firestore) collectGeneralCourses(ctx context.Context, query firestore.Query, limit, offset int, skipPagination bool) ([]types.CourseGeneralInfo, bool, error) {
	if !skipPagination && limit > 0 {
		query = query.Offset(offset).Limit(limit + 1)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var courses []types.CourseGeneralInfo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to iterate courses: %w", err)
		}
		var course types.CourseGeneralInfo
		if err := doc.DataTo(&course); err != nil {
			continue
		}
		courses = append(courses, course)
	}

	if skipPagination {
		return courses, false, nil
	}

	hasNext := limit > 0 && len(courses) > limit
	if hasNext {
		courses = courses[:limit]
	}
	return courses, hasNext, nil
}
