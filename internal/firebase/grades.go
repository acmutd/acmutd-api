package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

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

	query := c.CollectionGroup("records").
		Where("instructor_name_normalized", ">=", normalizedName).
		Where("instructor_name_normalized", "<=", normalizedName+"\uf8ff")

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
