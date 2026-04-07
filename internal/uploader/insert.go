package uploader

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
)

/*
Structure:

  - courses/{course_prefix}/numbers/{course_number}/sections/{section_address}

  - courses/{course_prefix}/numbers/{course_number}

  - terms/{term}/prefixes/{course_prefix}

    This mirrors the recommended Firestore layout for efficient collection group
    queries by term, course prefix, and course number while maintaining fast prefix
    lookups for each term.
*/
func (h *UploaderHandler) insertTermData(ctx context.Context, courses []*types.CourseGeneralInfo, term string, isLatestTerm bool) error {
	normalizedTerm := strings.ToLower(strings.TrimSpace(term))
	if normalizedTerm == "" {
		return fmt.Errorf("invalid term: %s", term)
	}

	writer := h.fbclient.Firestore().BulkWriter(ctx)
	defer writer.End()

	prefixes := h.insertCourses(writer, courses, normalizedTerm, isLatestTerm)

	h.insertTerm(writer, prefixes, normalizedTerm)

	return nil
}

func (h *UploaderHandler) insertCourses(writer *firestore.BulkWriter, courses []*types.CourseGeneralInfo, term string, isLatestTerm bool) map[string]string {
	prefixes := make(map[string]string)

	// For the latest term, update all general info fields alongside the grade distribution.
	// For older terms, only append the grade distribution to avoid overwriting newer metadata.
	mergePaths := []firestore.FieldPath{{"grades", term}}
	if isLatestTerm {
		mergePaths = append(mergePaths,
			firestore.FieldPath{"course_prefix"},
			firestore.FieldPath{"course_number"},
			firestore.FieldPath{"title"},
			firestore.FieldPath{"description"},
			firestore.FieldPath{"credit_hours"},
			firestore.FieldPath{"school"},
			firestore.FieldPath{"school_id"},
			firestore.FieldPath{"core"},
			firestore.FieldPath{"schedule_frequency"},
			firestore.FieldPath{"section_types"},
			firestore.FieldPath{"last_updated_term"},
		)
	}

	for _, course := range courses {
		if isLatestTerm {
			course.LastUpdatedTerm = term
		}
		cdoc := h.coursesCollection(course.Prefix, course.Number)
		writer.Set(cdoc, course, firestore.Merge(mergePaths...))

		for _, sec := range course.Sections {
			prepared, ok := prepareCourseForTerm(*sec, term)
			if !ok {
				continue
			}

			sdoc := cdoc.Collection("sections").Doc(prepared.SectionID)
			writer.Set(sdoc, prepared.Course)

			if _, exists := prefixes[prepared.PrefixID]; !exists {
				prefixes[prepared.PrefixID] = prepared.Course.Prefix
			}
		}
	}

	return prefixes
}

func (h *UploaderHandler) insertTerm(writer *firestore.BulkWriter, prefixes map[string]string, term string) {
	termDoc := h.fbclient.Firestore().Collection("terms").Doc(term)
	writer.Set(termDoc, map[string]any{
		"term":         term,
		"last_updated": time.Now(),
	}, firestore.MergeAll)

	for prefixID, originalPrefix := range prefixes {
		writer.Set(
			termDoc.Collection("prefixes").Doc(prefixID),
			map[string]any{
				"course_prefix":     originalPrefix,
				"normalized_prefix": prefixID,
			},
			firestore.MergeAll,
		)
	}
}

func (h *UploaderHandler) insertProfs(ctx context.Context, profs []*types.Professor) {
	writer := h.fbclient.Firestore().BulkWriter(ctx)
	defer writer.End()

	fs := h.fbclient.Firestore()
	profCollection := fs.Collection("professors")

	for _, prof := range profs {
		writer.Set(
			profCollection.Doc(prof.InstructorID),
			prof,
		)
	}
}

// sanitizeDocID sanitizes a value for use as a Firestore document ID
func sanitizeDocID(value string) string {
	sanitized := strings.TrimSpace(value)
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "")
	return sanitized
}

func ensureSectionDocID(course types.SectionDoc, term string) string {
	if section := sanitizeDocID(course.SectionAddress); section != "" {
		return strings.ToLower(section)
	}

	prefix := sanitizeDocID(strings.ToLower(strings.TrimSpace(course.Prefix)))
	number := sanitizeDocID(strings.ToLower(strings.TrimSpace(course.Number)))
	section := sanitizeDocID(strings.ToLower(course.Section))
	if section == "" {
		section = "000"
	}
	normalizedTerm := sanitizeDocID(strings.ToLower(strings.TrimSpace(term)))

	generated := fmt.Sprintf("%s%s.%s.%s", prefix, number, section, normalizedTerm)
	generated = strings.ReplaceAll(generated, "..", ".")
	generated = strings.Trim(generated, ".")

	return generated
}

type preparedCourse struct {
	Course    types.SectionDoc
	PrefixID  string
	NumberID  string
	SectionID string
}

func prepareCourseForTerm(course types.SectionDoc, term string) (preparedCourse, bool) {

	normalizedTerm := strings.ToLower(strings.TrimSpace(term))

	if normalizedTerm == "" {
		return preparedCourse{}, false
	}

	course.Term = normalizedTerm
	course.Prefix = strings.ToLower(strings.TrimSpace(course.Prefix))
	course.Number = strings.ToLower(strings.TrimSpace(course.Number))
	course.Section = strings.ToLower(strings.TrimSpace(course.Section))

	prefixID := sanitizeDocID(course.Prefix)
	numberID := sanitizeDocID(course.Number)
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

func (h *UploaderHandler) coursesCollection(prefixID, numberID string) *firestore.DocumentRef {
	return h.fbclient.Firestore().Collection("courses").
		Doc(prefixID).
		Collection("numbers").
		Doc(numberID)
}
