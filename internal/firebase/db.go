package firebase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Firestore struct {
	*firestore.Client
}

func NewFirestore(ctx context.Context, app *firebase.App) (*Firestore, error) {

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Firestore{
		Client: client,
	}, nil
}

func normalizeCoursePrefix(prefix string) string {
	return strings.ToLower(strings.TrimSpace(prefix))
}

func normalizeCourseNumber(number string) string {
	return strings.ToLower(strings.TrimSpace(number))
}

func normalizeTerm(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func sanitizeDocID(value string) string {
	sanitized := strings.TrimSpace(value)
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "")
	return sanitized
}

func ensureSectionDocID(course types.Course, term string) string {
	if section := sanitizeDocID(course.SectionAddress); section != "" {
		return strings.ToLower(section)
	}

	prefix := sanitizeDocID(normalizeCoursePrefix(course.CoursePrefix))
	number := sanitizeDocID(normalizeCourseNumber(course.CourseNumber))
	section := sanitizeDocID(strings.ToLower(course.Section))
	if section == "" {
		section = "000"
	}
	normalizedTerm := sanitizeDocID(normalizeTerm(term))

	generated := fmt.Sprintf("%s%s.%s.%s", prefix, number, section, normalizedTerm)
	generated = strings.ReplaceAll(generated, "..", ".")
	generated = strings.Trim(generated, ".")

	return generated
}

type preparedCourse struct {
	Course    types.Course
	PrefixID  string
	NumberID  string
	SectionID string
}

func prepareCourseForTerm(course types.Course, normalizedTerm string) (preparedCourse, bool) {
	if normalizedTerm == "" {
		return preparedCourse{}, false
	}

	course.Term = normalizedTerm
	course.CoursePrefix = normalizeCoursePrefix(course.CoursePrefix)
	course.CourseNumber = normalizeCourseNumber(course.CourseNumber)
	course.Section = strings.ToLower(strings.TrimSpace(course.Section))

	prefixID := sanitizeDocID(course.CoursePrefix)
	numberID := sanitizeDocID(course.CourseNumber)
	if prefixID == "" || numberID == "" {
		return preparedCourse{}, false
	}

	sectionID := ensureSectionDocID(course, normalizedTerm)
	course.SectionAddress = sectionID

	return preparedCourse{
		Course:    course,
		PrefixID:  prefixID,
		NumberID:  numberID,
		SectionID: sectionID,
	}, true
}

func (c *Firestore) sectionsCollection(prefixID, numberID string) *firestore.CollectionRef {
	return c.Collection("courses").
		Doc(prefixID).
		Collection("numbers").
		Doc(numberID).
		Collection("sections")
}

/*
Structure:

  - courses/{course_prefix}/numbers/{course_number}/sections/{section_address}

  - terms/{term}/prefixes/{course_prefix}

    This mirrors the recommended Firestore layout for efficient collection group
    queries by term, course prefix, and course number while maintaining fast prefix
    lookups for each term.
*/
func (c *Firestore) InsertClassesWithIndexes(ctx context.Context, courses []types.Course, term string) {
	writer := c.BulkWriter(ctx)
	defer writer.End()

	normalizedTerm := normalizeTerm(term)
	if normalizedTerm == "" {
		return
	}

	termDoc := c.Collection("terms").Doc(normalizedTerm)
	writer.Set(termDoc, map[string]any{
		"term":         normalizedTerm,
		"last_updated": time.Now(),
	}, firestore.MergeAll)

	prefixes := make(map[string]string)

	for _, course := range courses {
		prepared, ok := prepareCourseForTerm(course, normalizedTerm)
		if !ok {
			continue
		}

		doc := c.sectionsCollection(prepared.PrefixID, prepared.NumberID).Doc(prepared.SectionID)
		writer.Set(doc, prepared.Course)

		if _, exists := prefixes[prepared.PrefixID]; !exists {
			prefixes[prepared.PrefixID] = prepared.Course.CoursePrefix
		}
	}

	for prefixID, originalPrefix := range prefixes {
		writer.Set(
			termDoc.Collection("prefixes").Doc(prefixID),
			map[string]any{
				"course_prefix":     originalPrefix,
				"normalized_prefix": prefixID,
				"term":              normalizedTerm,
			},
			firestore.MergeAll,
		)
	}
}

func (c *Firestore) InsertTerms(ctx context.Context, terms []string) {
	writer := c.BulkWriter(ctx)
	defer writer.End()

	for _, term := range terms {
		normalized := normalizeTerm(term)
		if normalized == "" {
			continue
		}

		doc := c.Collection("terms").Doc(normalized)
		writer.Set(doc, map[string]any{
			"term": normalized,
		}, firestore.MergeAll)
	}
}

func (c *Firestore) QueryAllTerms(ctx context.Context, limit, offset int) ([]string, bool, error) {
	query := c.Collection("terms").OrderBy("term", firestore.Asc)
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit + 1)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var terms []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to get next term: %w", err)
		}

		data := doc.Data()
		if termValue, ok := data["term"].(string); ok && termValue != "" {
			terms = append(terms, termValue)
			continue
		}

		terms = append(terms, doc.Ref.ID)
	}

	sort.Strings(terms)

	hasNext := false
	if limit > 0 && len(terms) > limit {
		hasNext = true
		terms = terms[:limit]
	}

	return terms, hasNext, nil
}
func (c *Firestore) QueryByCourseNumber(ctx context.Context, term, coursePrefix, courseNumber string, limit, offset int) ([]types.Course, bool, error) {
	term = normalizeTerm(term)
	coursePrefix = normalizeCoursePrefix(coursePrefix)
	courseNumber = normalizeCourseNumber(courseNumber)
	if term == "" || coursePrefix == "" || courseNumber == "" {
		return []types.Course{}, false, nil
	}

	query := c.CollectionGroup("sections").
		Where("term", "==", term).
		Where("course_prefix", "==", coursePrefix).
		Where("course_number", "==", courseNumber)

	return c.collectCourses(ctx, query, limit, offset)
}

func (c *Firestore) QueryByCoursePrefix(ctx context.Context, term, coursePrefix string, limit, offset int) ([]types.Course, bool, error) {
	term = normalizeTerm(term)
	coursePrefix = normalizeCoursePrefix(coursePrefix)
	if term == "" || coursePrefix == "" {
		return []types.Course{}, false, nil
	}

	query := c.CollectionGroup("sections").
		Where("term", "==", term).
		Where("course_prefix", "==", coursePrefix)

	return c.collectCourses(ctx, query, limit, offset)
}

// GetAllCoursesByTerm returns all courses for a given term
func (c *Firestore) GetAllCoursesByTerm(ctx context.Context, term string, limit, offset int) ([]types.Course, bool, error) {
	term = normalizeTerm(term)
	if term == "" {
		return []types.Course{}, false, nil
	}

	query := c.CollectionGroup("sections").
		Where("term", "==", term)

	return c.collectCourses(ctx, query, limit, offset)
}

// QueryBySchool returns courses by school for a given term
func (c *Firestore) QueryBySchool(ctx context.Context, term, school string, limit, offset int) ([]types.Course, bool, error) {
	term = normalizeTerm(term)
	school = strings.TrimSpace(school)
	if term == "" || school == "" {
		return []types.Course{}, false, nil
	}

	query := c.CollectionGroup("sections").
		Where("term", "==", term).
		Where("school", "==", school)

	return c.collectCourses(ctx, query, limit, offset)
}

func (c *Firestore) GetCourseBySection(ctx context.Context, term string, prefix string, number string, section string) (*types.Course, error) {
	term = normalizeTerm(term)
	prefix = normalizeCoursePrefix(prefix)
	number = normalizeCourseNumber(number)
	section = strings.ToLower(strings.TrimSpace(section))

	id := fmt.Sprintf("%s%s.%s.%s", prefix, number, section, term)

	doc, err := c.Collection("courses").Doc(prefix).Collection("numbers").Doc(number).Collection("sections").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var course types.Course
	if err := doc.DataTo(&course); err != nil {
		return nil, err
	}

	return &course, nil
}

func (c *Firestore) collectCourses(ctx context.Context, query firestore.Query, limit, offset int) ([]types.Course, bool, error) {
	if limit > 0 {
		query = query.Offset(offset).Limit(limit + 1)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var courses []types.Course
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to get next document: %w", err)
		}

		var course types.Course
		if err := doc.DataTo(&course); err != nil {
			continue
		}
		courses = append(courses, course)
	}

	hasNext := false
	if limit > 0 && len(courses) > limit {
		hasNext = true
		courses = courses[:limit]
	}

	return courses, hasNext, nil
}

// SearchCourses searches courses by title, topic, or instructor name
func (c *Firestore) SearchCourses(ctx context.Context, term, searchQuery string, limit, offset int) ([]types.Course, bool, error) {
	// TODO: Figure out a nicer way to do this
	normalizedTerm := normalizeTerm(term)
	courses, _, err := c.GetAllCoursesByTerm(ctx, normalizedTerm, 0, 0)
	if err != nil {
		return nil, false, err
	}

	query := strings.ToLower(strings.TrimSpace(searchQuery))
	if query == "" {
		if limit <= 0 {
			return courses, false, nil
		}
		start := offset
		if start >= len(courses) {
			return []types.Course{}, false, nil
		}
		end := offset + limit
		if end > len(courses) {
			end = len(courses)
		}
		hasNext := end < len(courses)
		return courses[start:end], hasNext, nil
	}

	var filteredCourses []types.Course

	for _, course := range courses {
		title := strings.ToLower(course.Title)
		topic := strings.ToLower(course.Topic)
		instructors := strings.ToLower(course.Instructors)

		// Search in title, topic, and instructors
		if strings.Contains(title, query) ||
			strings.Contains(topic, query) ||
			strings.Contains(instructors, query) {
			filteredCourses = append(filteredCourses, course)
		}
	}

	if limit <= 0 {
		return filteredCourses, false, nil
	}

	start := offset
	if start >= len(filteredCourses) {
		return []types.Course{}, false, nil
	}

	end := offset + limit
	if end > len(filteredCourses) {
		end = len(filteredCourses)
	}

	hasNext := end < len(filteredCourses)

	return filteredCourses[start:end], hasNext, nil
}

// GetSchoolsByTerm returns all schools for a given term
func (c *Firestore) GetSchoolsByTerm(ctx context.Context, term string) ([]string, error) {
	term = normalizeTerm(term)
	if term == "" {
		return nil, nil
	}

	iter := c.Collection("terms").Doc(term).Collection("prefixes").Documents(ctx)
	defer iter.Stop()

	var prefixes []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate prefixes: %w", err)
		}

		data := doc.Data()
		if prefix, ok := data["course_prefix"].(string); ok && strings.TrimSpace(prefix) != "" {
			prefixes = append(prefixes, prefix)
			continue
		}

		prefixes = append(prefixes, doc.Ref.ID)
	}

	if len(prefixes) == 0 {
		fallbackQuery := c.CollectionGroup("sections").Where("term", "==", term)
		fallbackIter := fallbackQuery.Documents(ctx)
		defer fallbackIter.Stop()

		uniquePrefixes := make(map[string]struct{})
		for {
			doc, err := fallbackIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to iterate fallback prefixes: %w", err)
			}

			var course types.Course
			if err := doc.DataTo(&course); err != nil {
				return nil, fmt.Errorf("failed to parse fallback course: %w", err)
			}

			prefix := strings.TrimSpace(course.CoursePrefix)
			if prefix == "" {
				continue
			}
			uniquePrefixes[prefix] = struct{}{}
		}

		for prefix := range uniquePrefixes {
			prefixes = append(prefixes, prefix)
		}
	}

	if len(prefixes) == 0 {
		return nil, nil
	}

	unique := make(map[string]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		trimmed := strings.TrimSpace(prefix)
		if trimmed == "" {
			continue
		}
		unique[trimmed] = struct{}{}
	}

	prefixes = prefixes[:0]
	for prefix := range unique {
		prefixes = append(prefixes, prefix)
	}

	sort.Strings(prefixes)

	return prefixes, nil
}

func (c *Firestore) GetProfessorById(ctx context.Context, id string) (*types.Professor, error) {
	doc, err := c.Collection("professors").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var professor types.Professor
	if err := doc.DataTo(&professor); err != nil {
		return nil, err
	}

	return &professor, nil
}

func (c *Firestore) GetProfessorsByName(ctx context.Context, name string, limit, offset int) ([]types.Professor, bool, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return []types.Professor{}, false, nil
	}

	query := c.Collection("professors").Where("normalized_coursebook_name", "==", normalizedName)

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit + 1)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()

	var professors []types.Professor
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to get next professor: %w", err)
		}

		var professor types.Professor
		if err := doc.DataTo(&professor); err != nil {
			continue
		}
		professors = append(professors, professor)
	}
	hasNext := false
	if limit > 0 && len(professors) > limit {
		hasNext = true
		professors = professors[:limit]
	}

	return professors, hasNext, nil
}

func (c *Firestore) GetGradesByPrefix(ctx context.Context, prefix string, limit, offset int) ([]types.Grades, bool, error) {
	query := c.CollectionGroup("records").Where("course_prefix", "==", prefix)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) GetGradesByPrefixAndNumber(ctx context.Context, prefix, number string, limit, offset int) ([]types.Grades, bool, error) {
	records := c.Collection("grades").Doc(prefix).Collection("courses").Doc(number).Collection("records")
	return c.collectGrades(ctx, records.Query, limit, offset)
}

func (c *Firestore) GetGradesByPrefixAndTerm(ctx context.Context, prefix, term string, limit, offset int) ([]types.Grades, bool, error) {
	query := c.CollectionGroup("records").Where("course_prefix", "==", prefix).Where("term", "==", term)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) GetGradesByNumberAndTerm(ctx context.Context, term string, prefix, number string, limit, offset int) ([]types.Grades, bool, error) {
	query := c.Collection("grades").Doc(prefix).Collection("courses").Doc(number).Collection("records").Where("term", "==", term)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) GetGradesBySection(ctx context.Context, term string, prefix, number string, section string) (*types.Grades, error) {
	id := fmt.Sprintf("%s%s.%s.%s", prefix, number, section, term)

	doc, err := c.Collection("grades").Doc(prefix).Collection("courses").Doc(number).Collection("records").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var grades types.Grades
	if err := doc.DataTo(&grades); err != nil {
		return nil, err
	}

	return &grades, nil
}

func (c *Firestore) GetGradesByProfId(ctx context.Context, profId string, limit, offset int) ([]types.Grades, bool, error) {
	query := c.CollectionGroup("records").Where("instructor_id", "==", profId)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) GetGradesByProfName(ctx context.Context, profName string, limit, offset int) ([]types.Grades, bool, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(profName))
	if normalizedName == "" {
		return []types.Grades{}, false, nil
	}

	query := c.CollectionGroup("records").Where("instructor_name_normalized", "==", normalizedName)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) GetGradesByTerm(ctx context.Context, term string, limit, offset int) ([]types.Grades, bool, error) {
	query := c.CollectionGroup("records").Where("term", "==", term)
	return c.collectGrades(ctx, query, limit, offset)
}

func (c *Firestore) collectGrades(ctx context.Context, query firestore.Query, limit, offset int) ([]types.Grades, bool, error) {
	if limit > 0 {
		query = query.Offset(offset).Limit(limit + 1)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var grades []types.Grades
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to get next grade: %w", err)
		}

		var grade types.Grades
		if err := doc.DataTo(&grade); err != nil {
			continue
		}
		grades = append(grades, grade)
	}

	hasNext := false
	if limit > 0 && len(grades) > limit {
		hasNext = true
		grades = grades[:limit]
	}

	return grades, hasNext, nil
}

func (c *Firestore) GenerateAPIKey(
	ctx context.Context,
	rateLimit int,
	windowSeconds int,
	isAdmin bool,
	expiresAt time.Time,
) (string, error) {
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	key := hex.EncodeToString(keyBytes)

	apiKey := types.APIKey{
		Key:           key,
		RateLimit:     rateLimit,
		WindowSeconds: windowSeconds,
		IsAdmin:       isAdmin,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		UsageCount:    0,
	}

	_, err := c.Collection("api_keys").Doc(key).Set(ctx, apiKey)
	return key, err
}

// ValidateAPIKey with expiration check
func (c *Firestore) ValidateAPIKey(ctx context.Context, key string) (*types.APIKey, error) {
	doc, err := c.Collection("api_keys").Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, err
	}

	// We don't need to check expiration here because it's checked in the middleware
	return &apiKey, nil
}

// UpdateKeyUsage updates last used and usage count
func (c *Firestore) UpdateKeyUsage(ctx context.Context, key string) error {
	_, err := c.Collection("api_keys").Doc(key).Update(ctx, []firestore.Update{
		{Path: "usage_count", Value: firestore.Increment(1)},
	})
	return err
}

func (c *Firestore) GetAPIKey(ctx context.Context, key string) (*types.APIKey, error) {
	doc, err := c.Collection("api_keys").Doc(key).Get(ctx)
	if err != nil {
		return nil, err
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// DeleteAllAdminKeys deletes all existing admin keys from Firebase
func (c *Firestore) DeleteAllAdminKeys(ctx context.Context) (returnedErr error) {
	// Query all documents in api_keys collection where is_admin is true
	iter := c.Collection("api_keys").Where("is_admin", "==", true).Documents(ctx)
	defer iter.Stop()

	batch := c.BulkWriter(ctx)
	defer batch.End()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if status.Code(err) == codes.NotFound {
				break
			}
			return fmt.Errorf("failed to iterate admin keys: %w", err)
		}

		batch.Delete(doc.Ref)
	}

	return nil
}

// GenerateAdminAPIKey generates a new admin API key with the "admin-" prefix
func (c *Firestore) GenerateAdminAPIKey(ctx context.Context) (string, error) {
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	baseKey := hex.EncodeToString(keyBytes)
	adminKey := "admin-" + baseKey

	apiKey := types.APIKey{
		Key:           adminKey,
		RateLimit:     0, // No rate limit for admin
		WindowSeconds: 0,
		IsAdmin:       true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Time{}, // Never expires
		UsageCount:    0,
	}

	// Store with the full admin key as the document ID
	_, err := c.Collection("api_keys").Doc(adminKey).Set(ctx, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to store admin key: %w", err)
	}

	return adminKey, nil
}
