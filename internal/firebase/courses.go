package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

func (c *Firestore) QueryCourses(ctx context.Context, q types.CourseQuery) ([]types.Course, bool, error) {
	term := strings.ToLower(strings.TrimSpace(q.Term))
	if term == "" {
		return []types.Course{}, false, nil
	}

	prefix := strings.ToLower(strings.TrimSpace(q.CoursePrefix))
	number := strings.ToLower(strings.TrimSpace(q.CourseNumber))
	section := strings.ToLower(strings.TrimSpace(q.Section))
	school := strings.ToLower(strings.TrimSpace(q.School))
	search := strings.ToLower(strings.TrimSpace(q.Search))
	instructor := strings.ToLower(strings.TrimSpace(q.Instructor))
	instructorID := strings.ToLower(strings.TrimSpace(q.InstructorID))
	days := strings.ToLower(strings.TrimSpace(q.Days))
	times := strings.TrimSpace(q.Times)
	times12h := strings.TrimSpace(q.Times12h)
	location := strings.ToLower(strings.TrimSpace(q.Location))

	query := c.CollectionGroup("sections").Where("term", "==", term)

	if prefix != "" {
		query = query.Where("course_prefix", "==", prefix)
	}
	if number != "" {
		query = query.Where("course_number", "==", number)
	}
	if section != "" {
		// our firestore data sometimes has an extra space after the section for some reason
		// this will query for both section and section with extra whitespace
		query = query.Where("section", "in", []string{section, section + " "})
	}
	if school != "" {
		query = query.Where("school", "==", school)
	}

	// Check if any manual filtering is needed (substring matching not supported by Firestore)
	needsManualFilter := search != "" || instructor != "" || instructorID != "" || days != "" || times != "" || times12h != "" || location != ""

	courses, hasNext, err := c.collectCourses(ctx, query, q.Limit, q.Offset, needsManualFilter)
	if err != nil {
		return nil, false, err
	}

	// Apply manual filters if needed
	if needsManualFilter {
		var filteredCourses []types.Course
		for _, course := range courses {
			// Apply search filter
			if search != "" {
				title := strings.ToLower(course.Title)
				topic := strings.ToLower(course.Topic)
				instructors := strings.ToLower(course.Instructors)

				if !strings.Contains(title, search) &&
					!strings.Contains(topic, search) &&
					!strings.Contains(instructors, search) {
					continue
				}
			}

			// Apply instructor filter
			if instructor != "" && !strings.Contains(strings.ToLower(course.Instructors), instructor) {
				continue
			}

			// Apply instructor_id filter
			if instructorID != "" && !strings.Contains(strings.ToLower(course.InstructorIDs), instructorID) {
				continue
			}

			// Apply days filter
			if days != "" && !strings.Contains(strings.ToLower(course.Days), days) {
				continue
			}

			// Apply times filter
			if times != "" && !strings.Contains(course.Times, times) {
				continue
			}

			// Apply times_12h filter
			if times12h != "" && !strings.Contains(course.Times12h, times12h) {
				continue
			}

			// Apply location filter
			// sometimes location is "building_number" and sometimes is "building number" in firestore
			// this checks for matches of both
			if location != "" && !strings.Contains(strings.ReplaceAll(strings.ToLower(course.Location), "_", " "), strings.ReplaceAll(location, "_", " ")) {
				continue
			}

			filteredCourses = append(filteredCourses, course)
		}
		courses = filteredCourses

		// Apply pagination for manually filtered results
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

		hasNext = end < len(courses)
		return courses[start:end], hasNext, nil
	}

	return courses, hasNext, nil
}

func (c *Firestore) collectCourses(ctx context.Context, query firestore.Query, limit, offset int, skipPagination bool) ([]types.Course, bool, error) {
	// If pagination not skipped, apply pagination at the query level
	if !skipPagination && limit > 0 {
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

	// If pagination was skipped, return all results for further filtering
	if skipPagination {
		return courses, false, nil
	}

	// Check if there are more results
	hasNext := false
	if limit > 0 && len(courses) > limit {
		hasNext = true
		courses = courses[:limit]
	}

	return courses, hasNext, nil
}
