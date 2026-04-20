package firebase

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateOrUpdateUser upserts a user record on every Firebase Auth login.
// On first login the user is created with approval_status "approved".
// On subsequent logins only display_name and last_login_at are refreshed.
func (c *Firestore) CreateOrUpdateUser(ctx context.Context, uid string, claims map[string]any) (*types.User, error) {
	email, _ := claims["email"].(string)
	displayName, _ := claims["name"].(string)
	now := time.Now()

	ref := c.Collection("users").Doc(uid)
	doc, err := ref.Get(ctx)

	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}

	if status.Code(err) == codes.NotFound {
		user := types.User{
			UID:            uid,
			Email:          email,
			DisplayName:    displayName,
			IsAdmin:        false,
			ApprovalStatus: "approved",
			CreatedAt:      now,
			LastLoginAt:    now,
		}
		if _, setErr := ref.Set(ctx, user); setErr != nil {
			return nil, setErr
		}
		return &user, nil
	}

	if _, updateErr := ref.Update(ctx, []firestore.Update{
		{Path: "display_name", Value: displayName},
		{Path: "last_login_at", Value: now},
	}); updateErr != nil {
		return nil, updateErr
	}

	var user types.User
	if dataErr := doc.DataTo(&user); dataErr != nil {
		return nil, dataErr
	}
	user.DisplayName = displayName
	user.LastLoginAt = now
	return &user, nil
}

// GetUser retrieves a single user by UID.
func (c *Firestore) GetUser(ctx context.Context, uid string) (*types.User, error) {
	doc, err := c.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	var user types.User
	return &user, doc.DataTo(&user)
}

// ListUsers returns all users ordered by created_at descending.
func (c *Firestore) ListUsers(ctx context.Context) ([]types.User, error) {
	iter := c.Collection("users").OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var users []types.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var u types.User
		if err := doc.DataTo(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserByEmail returns the user with the given email address, or nil if not found.
func (c *Firestore) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	iter := c.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var user types.User
	return &user, doc.DataTo(&user)
}

// SetUserBanStatus sets approval_status to "banned" or "approved".
// When banning, all api_keys owned by the user are also set to "inactive".
func (c *Firestore) SetUserBanStatus(ctx context.Context, uid string, banned bool) error {
	newStatus := "approved"
	if banned {
		newStatus = "banned"
	}

	if _, err := c.Collection("users").Doc(uid).Update(ctx, []firestore.Update{
		{Path: "approval_status", Value: newStatus},
	}); err != nil {
		return err
	}

	if !banned {
		return nil
	}

	// Deactivate all keys belonging to the user
	iter := c.Collection("api_keys").Where("user_id", "==", uid).Documents(ctx)
	defer iter.Stop()

	batch := c.BulkWriter(ctx)
	defer batch.End()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		batch.Update(doc.Ref, []firestore.Update{{Path: "status", Value: "inactive"}})
	}

	return nil
}
