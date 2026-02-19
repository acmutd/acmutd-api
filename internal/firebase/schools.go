package firebase

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

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
