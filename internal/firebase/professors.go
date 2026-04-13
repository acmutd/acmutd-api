package firebase

import (
	"context"
	"fmt"

	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

func (c *Firestore) GetProfessorById(ctx context.Context, id string) (*types.Professor, error) {
	key := cacheKey("professors", "id", id)
	if cached, found := c.Cache.Get(key); found {
		if p, ok := cached.(types.Professor); ok {
			return &p, nil
		}
	}

	doc, err := c.Collection("professors").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var professor types.Professor
	if err := doc.DataTo(&professor); err != nil {
		return nil, err
	}

	c.Cache.Set(key, professor, c.TTL.Professors)

	return &professor, nil
}

func (c *Firestore) GetProfessorsByName(ctx context.Context, name string, limit, offset int) ([]types.Professor, bool, error) {
	key := cacheKey("professors", "name", name)
	if cached, found := c.Cache.Get(key); found {
		if p, ok := cached.(cachedResult[types.Professor]); ok {
			return p.Items, p.HasNext, nil
		}
	}

	query := c.Collection("professors").
		Where("normalized_coursebook_name", ">=", name).
		Where("normalized_coursebook_name", "<=", name+"\uf8ff")

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

	c.Cache.Set(key, cachedResult[types.Professor]{Items: professors, HasNext: hasNext}, c.TTL.Professors)

	return professors, hasNext, nil
}
