package firebase

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

// GetDailyStats returns usage statistics for a specific user, ordered by date descending.
func (c *Firestore) GetDailyStats(ctx context.Context, userID string) ([]types.DailyStat, error) {
	iter := c.Collection("daily_stats").
		Where("user_id", "==", userID).
		OrderBy("date", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	return collectDailyStats(iter)
}

// ListAllDailyStats returns all daily stats ordered by date descending.
func (c *Firestore) ListAllDailyStats(ctx context.Context) ([]types.DailyStat, error) {
	iter := c.Collection("daily_stats").
		OrderBy("date", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	return collectDailyStats(iter)
}

func collectDailyStats(iter *firestore.DocumentIterator) ([]types.DailyStat, error) {
	var stats []types.DailyStat
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var s types.DailyStat
		if err := doc.DataTo(&s); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetRecentRequests returns recent request events for a specific user, ordered by date_time descending.
func (c *Firestore) GetRecentRequests(ctx context.Context, userID string) ([]types.RequestEvent, error) {
	iter := c.Collection("request_events").
		Where("user_id", "==", userID).
		OrderBy("date_time", firestore.Desc).
		Limit(100).
		Documents(ctx)
	defer iter.Stop()

	return collectRequestEvents(iter)
}

// ListAllRecentRequests returns all request events ordered by date_time descending.
func (c *Firestore) ListAllRecentRequests(ctx context.Context) ([]types.RequestEvent, error) {
	iter := c.Collection("request_events").
		OrderBy("date_time", firestore.Desc).
		Limit(500).
		Documents(ctx)
	defer iter.Stop()

	return collectRequestEvents(iter)
}

func collectRequestEvents(iter *firestore.DocumentIterator) ([]types.RequestEvent, error) {
	var events []types.RequestEvent
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var e types.RequestEvent
		if err := doc.DataTo(&e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
