package types

import "time"

type ScraperLog struct {
	LogID           string    `firestore:"log_id" json:"logId"`
	StartedAt       time.Time `firestore:"started_at" json:"startedAt"`
	CompletedAt     time.Time `firestore:"completed_at" json:"completedAt"`
	Status          string    `firestore:"status" json:"status"` // "running" | "success" | "error"
	Semester        string    `firestore:"semester" json:"semester"`
	CoursesWritten  int       `firestore:"courses_written" json:"coursesWritten"`
	SectionsWritten int       `firestore:"sections_written" json:"sectionsWritten"`
	MergeMatchRate  float64   `firestore:"merge_match_rate" json:"mergeMatchRate"`
	TriggeredBy     string    `firestore:"triggered_by" json:"triggeredBy"`
	ErrorMessage    string    `firestore:"error_message" json:"errorMessage,omitempty"`
}

type CronLog struct {
	LogID           string    `firestore:"log_id" json:"logId"`
	RunAt           time.Time `firestore:"run_at" json:"runAt"`
	JobName         string    `firestore:"job_name" json:"jobName"`
	Status          string    `firestore:"status" json:"status"` // "success" | "error"
	TokensChecked   int       `firestore:"tokens_checked" json:"tokensChecked"`
	ValidTokenCount int       `firestore:"valid_token_count" json:"validTokenCount"`
	Decision        string    `firestore:"decision" json:"decision"` // "start" | "stop" | "no-op"
	TriggeredAction bool      `firestore:"triggered_action" json:"triggeredAction"`
	ServerLogID     string    `firestore:"server_log_id" json:"serverLogId"`
	ErrorMessage    string    `firestore:"error_message" json:"errorMessage"`
	DurationMs      int       `firestore:"duration_ms" json:"durationMs"`
}

type ServerStateLog struct {
	LogID         string    `firestore:"log_id" json:"logId"`
	Timestamp     time.Time `firestore:"timestamp" json:"timestamp"`
	Action        string    `firestore:"action" json:"action"` // "start" | "stop" | "resize" | "hackathon_enable" | "hackathon_disable"
	Status        string    `firestore:"status" json:"status"` // "success" | "error"
	TriggerSource string    `firestore:"trigger_source" json:"triggerSource"` // "admin_dashboard" | "cron" | "system"
	ActorType     string    `firestore:"actor_type" json:"actorType"`         // "user" | "system"
	ActorID       string    `firestore:"actor_id" json:"actorId"`
	ActorEmail    string    `firestore:"actor_email" json:"actorEmail"`
	Reason        string    `firestore:"reason" json:"reason"`
	RequestID     string    `firestore:"request_id" json:"requestId"`
	CronLogID     string    `firestore:"cron_log_id" json:"cronLogId"`
	PreviousState string    `firestore:"previous_state" json:"previousState"`
	NewState      string    `firestore:"new_state" json:"newState"`
	PreviousType  string    `firestore:"previous_type" json:"previousType"`
	NewType       string    `firestore:"new_type" json:"newType"`
	InstanceID    string    `firestore:"instance_id" json:"instanceId"`
	PublicIP      string    `firestore:"public_ip" json:"publicIp"`
	ErrorMessage  string    `firestore:"error_message" json:"errorMessage"`
	DurationMs    int       `firestore:"duration_ms" json:"durationMs"`
}

type PromotionLog struct {
	LogID         string    `firestore:"log_id" json:"logId"`
	PromotedAt    time.Time `firestore:"promoted_at" json:"promotedAt"`
	DocumentCount int       `firestore:"document_count" json:"documentCount"`
	PromotedBy    string    `firestore:"promoted_by" json:"promotedBy"`
	FromEnv       string    `firestore:"from_env" json:"fromEnv"`
	ToEnv         string    `firestore:"to_env" json:"toEnv"`
	Notes         string    `firestore:"notes" json:"notes,omitempty"`
}
