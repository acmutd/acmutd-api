package firebase

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
)

// ListScraperLogs returns all scraper logs ordered by started_at descending.
func (c *Firestore) ListScraperLogs(ctx context.Context) ([]types.ScraperLog, error) {
	iter := c.Collection("scraper_logs").
		OrderBy("started_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var logs []types.ScraperLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var l types.ScraperLog
		if err := doc.DataTo(&l); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ListCronLogs returns all cron logs ordered by run_at descending.
func (c *Firestore) ListCronLogs(ctx context.Context) ([]types.CronLog, error) {
	iter := c.Collection("cron_logs").
		OrderBy("run_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var logs []types.CronLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var l types.CronLog
		if err := doc.DataTo(&l); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ListServerStateLogs returns all server state logs ordered by timestamp descending.
func (c *Firestore) ListServerStateLogs(ctx context.Context) ([]types.ServerStateLog, error) {
	iter := c.Collection("server_state_logs").
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var logs []types.ServerStateLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var l types.ServerStateLog
		if err := doc.DataTo(&l); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ListPromotionLogs returns all promotion logs ordered by promoted_at descending.
func (c *Firestore) ListPromotionLogs(ctx context.Context) ([]types.PromotionLog, error) {
	iter := c.Collection("promotion_logs").
		OrderBy("promoted_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var logs []types.PromotionLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var l types.PromotionLog
		if err := doc.DataTo(&l); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// CreateServerStateLog writes a new server state log entry.
func (c *Firestore) CreateServerStateLog(ctx context.Context, log types.ServerStateLog) error {
	_, err := c.Collection("server_state_logs").Doc(log.LogID).Set(ctx, log)
	return err
}
