package types

import "time"

type User struct {
	UID            string    `firestore:"uid" json:"uid"`
	Email          string    `firestore:"email" json:"email"`
	DisplayName    string    `firestore:"display_name" json:"displayName"`
	IsAdmin        bool      `firestore:"is_admin" json:"isAdmin"`
	ApprovalStatus string    `firestore:"approval_status" json:"approvalStatus"` // "approved" | "banned"
	CreatedAt      time.Time `firestore:"created_at" json:"createdAt"`
	LastLoginAt    time.Time `firestore:"last_login_at" json:"lastLoginAt"`
}
