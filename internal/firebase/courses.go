package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QuerySections queries SectionDoc documents.
// term is required. When both prefix and number are provided, the query is scoped directly to
// courses/{prefix}/numbers/{number}/sections. When only prefix is provided, a collection group
// query filtered by prefix is used.
// instructor and days both use array-contains and cannot be combined in the same query.
func (c *Firestore) QuerySections(ctx context.Context, q types.SectionQuery) ([]types.SectionDoc, bool, error) {
	term := strings.ToLower(strings.TrimSpace(q.Term))
	if term == "" {
		return []types.SectionDoc{}, false, nil
	}

	prefix := strings.ToLower(strings.TrimSpace(q.Prefix))
	number := strings.ToLower(strings.TrimSpace(q.Number))
	instructor := strings.TrimSpace(q.Instructor)
	days := strings.TrimSpace(q.Days)

	var query firestore.Query
	if prefix != "" && number != "" {
		colRef := c.Collection("courses").Doc(prefix).Collection("numbers").Doc(number).Collection("sections")
		query = colRef.Where("term", "==", term)
	} else {
		query = c.CollectionGroup("sections").Where("term", "==", term)
		if prefix != "" {
			query = query.Where("prefix", "==", prefix)
		}
	}

	if instructor != "" {
		query = query.Where("instructors", "array-contains", instructor)
	} else if days != "" {
		query = query.Where("days", "array-contains", days)
	}

	return c.collectSections(ctx, query, q.Limit, q.Offset, false)
}

// GetSection retrieves a single SectionDoc by its section_address (e.g., "cs2305.001.23f").
// Returns nil, nil if not found.
func (c *Firestore) GetSection(ctx context.Context, sectionAddress string) (*types.SectionDoc, error) {
	sectionAddress = strings.ToLower(strings.TrimSpace(sectionAddress))
	if sectionAddress == "" {
		return nil, nil
	}

	iter := c.CollectionGroup("sections").
		Where("section_address", "==", sectionAddress).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	var s types.SectionDoc
	if err := doc.DataTo(&s); err != nil {
		return nil, fmt.Errorf("failed to parse section: %w", err)
	}
	return &s, nil
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

// GetSectionByParams retrieves a single SectionDoc via a direct document path.
// The section_address document ID is constructed as "{prefix}{number}.{section}.{term}".
// Returns nil, nil if not found.
func (c *Firestore) GetSectionByParams(ctx context.Context, prefix, number, term, section string) (*types.SectionDoc, error) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	number = strings.ToLower(strings.TrimSpace(number))
	term = strings.ToLower(strings.TrimSpace(term))
	section = strings.ToLower(strings.TrimSpace(section))

	sectionAddress := prefix + number + "." + section + "." + term

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
	return &s, nil
}

// QueryGeneralCourses queries CourseGeneralInfo documents.
// If prefix is provided, the query is scoped to courses/{prefix}/numbers.
// Otherwise a collection group query across all "numbers" subcollections is used.
func (c *Firestore) QueryGeneralCourses(ctx context.Context, q types.GeneralCourseQuery) ([]types.CourseGeneralInfo, bool, error) {
	prefix := strings.ToLower(strings.TrimSpace(q.Prefix))
	search := strings.ToLower(strings.TrimSpace(q.Search))

	var query firestore.Query
	if prefix != "" {
		query = c.Collection("courses").Doc(prefix).Collection("numbers").Query
	} else {
		query = c.CollectionGroup("numbers").Query
	}

	needsManualFilter := search != ""

	courses, hasNext, err := c.collectGeneralCourses(ctx, query, q.Limit, q.Offset, needsManualFilter)
	if err != nil {
		return nil, false, err
	}

	if !needsManualFilter {
		return courses, hasNext, nil
	}

	filtered := courses[:0]
	for _, course := range courses {
		title := strings.ToLower(course.Title)
		desc := strings.ToLower(course.Description)
		if !strings.Contains(title, search) && !strings.Contains(desc, search) {
			continue
		}
		filtered = append(filtered, course)
	}

	if q.Limit <= 0 {
		return filtered, false, nil
	}

	start := q.Offset
	if start >= len(filtered) {
		return []types.CourseGeneralInfo{}, false, nil
	}
	end := q.Offset + q.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], end < len(filtered), nil
}

// GetGeneralCourse retrieves a single CourseGeneralInfo at courses/{prefix}/numbers/{number}.
// Returns nil, nil if not found.
func (c *Firestore) GetGeneralCourse(ctx context.Context, prefix, number string) (*types.CourseGeneralInfo, error) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	number = strings.ToLower(strings.TrimSpace(number))
	if prefix == "" || number == "" {
		return nil, nil
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
