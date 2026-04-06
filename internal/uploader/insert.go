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

  - terms/{term}/prefixes/{course_prefix}

    This mirrors the recommended Firestore layout for efficient collection group
    queries by term, course prefix, and course number while maintaining fast prefix
    lookups for each term.
*/
func (h *UploaderHandler) InsertClassesWithIndexes(ctx context.Context, sections []*types.SectionDoc, term string) {
	fs := h.fbclient.Firestore()
	writer := fs.BulkWriter(ctx)
	defer writer.End()

	normalizedTerm := strings.ToLower(strings.TrimSpace(term))
	if normalizedTerm == "" {
		return
	}

	termDoc := fs.Collection("terms").Doc(normalizedTerm)
	writer.Set(termDoc, map[string]any{
		"term":         normalizedTerm,
		"last_updated": time.Now(),
	}, firestore.MergeAll)

	prefixes := make(map[string]string)

	for _, sec := range sections {
		prepared, ok := prepareCourseForTerm(*sec, normalizedTerm)
		if !ok {
			continue
		}

		doc := h.sectionsCollection(prepared.PrefixID, prepared.NumberID).Doc(prepared.SectionID)
		writer.Set(doc, prepared.Course)

		if _, exists := prefixes[prepared.PrefixID]; !exists {
			prefixes[prepared.PrefixID] = prepared.Course.Prefix
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

func (h *UploaderHandler) sectionsCollection(prefixID, numberID string) *firestore.CollectionRef {
	return h.fbclient.Firestore().Collection("courses").
		Doc(prefixID).
		Collection("numbers").
		Doc(numberID).
		Collection("sections")
}
