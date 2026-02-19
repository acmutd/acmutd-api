package firebase

import (
	"context"
	"fmt"
	"strings"

	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

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

	query := c.Collection("professors").
		Where("normalized_coursebook_name", ">=", normalizedName).
		Where("normalized_coursebook_name", "<=", normalizedName+"\uf8ff")

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
