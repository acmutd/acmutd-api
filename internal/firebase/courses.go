package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

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

func (c *Firestore) QueryCourses(ctx context.Context, q types.CourseQuery) ([]types.Course, bool, error) {
	term := normalizeTerm(q.Term)
	if term == "" {
		return []types.Course{}, false, nil
	}

	prefix := normalizeCoursePrefix(q.CoursePrefix)
	number := normalizeCourseNumber(q.CourseNumber)
	section := normalizeSection(q.Section)
	school := normalizeSchool(q.School)

	query := c.CollectionGroup("sections").Where("term", "==", term)

	if prefix != "" {
		query = query.Where("course_prefix", "==", prefix)
	}
	if number != "" {
		query = query.Where("course_number", "==", number)
	}
	if section != "" {
		// our firestore data adds an extra space after the section for some reason
		query = query.Where("section", "==", section+" ")
	}
	if school != "" {
		query = query.Where("school", "==", school)
	}

	return c.collectCourses(ctx, query, q.Limit, q.Offset)
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
func (c *Firestore) SearchCourses(ctx context.Context, q types.CourseQuery) ([]types.Course, bool, error) {
	// TODO: Figure out a nicer way to do this
	// Firestore doesn't implement full-text search, so we currently need to
	// query all courses by term and manually perform a search ourselves
	courses, _, err := c.QueryCourses(ctx, types.CourseQuery{Term: q.Term})
	if err != nil {
		return nil, false, err
	}

	search := strings.ToLower(strings.TrimSpace(q.Search))
	if search == "" {
		if q.Limit <= 0 {
			return courses, false, nil
		}
		start := q.Offset
		if start >= len(courses) {
			return []types.Course{}, false, nil
		}
		end := q.Offset + q.Limit
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
		if strings.Contains(title, search) ||
			strings.Contains(topic, search) ||
			strings.Contains(instructors, search) {
			filteredCourses = append(filteredCourses, course)
		}
	}

	if q.Limit <= 0 {
		return filteredCourses, false, nil
	}

	start := q.Offset
	if start >= len(filteredCourses) {
		return []types.Course{}, false, nil
	}

	end := q.Offset + q.Limit
	if end > len(filteredCourses) {
		end = len(filteredCourses)
	}

	hasNext := end < len(filteredCourses)

	return filteredCourses[start:end], hasNext, nil
}
