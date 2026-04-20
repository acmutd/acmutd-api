package types

import "time"

type EndpointBreakdown struct {
	Endpoint string `firestore:"endpoint" json:"endpoint"`
	Count    int    `firestore:"count" json:"count"`
}

type DailyStat struct {
	StatID         string         `firestore:"stat_id" json:"statId"`
	KeyID          string         `firestore:"key_id" json:"keyId"`
	UserID         string         `firestore:"user_id" json:"userId"`
	Date           string         `firestore:"date" json:"date"` // "YYYY-MM-DD"
	TotalRequests  int            `firestore:"total_requests" json:"totalRequests"`
	SuccessCount   int            `firestore:"success_count" json:"successCount"`
	ErrorCount     int            `firestore:"error_count" json:"errorCount"`
	EndpointCounts map[string]int `firestore:"endpoint_counts" json:"endpointCounts"`
}

type RequestEvent struct {
	ID         string    `firestore:"id" json:"id"`
	UserID     string    `firestore:"user_id" json:"userId"`
	KeyID      string    `firestore:"key_id" json:"keyId"`
	DateTime   time.Time `firestore:"date_time" json:"dateTime"`
	Endpoint   string    `firestore:"endpoint" json:"endpoint"`
	Method     string    `firestore:"method" json:"method"`
	StatusCode int       `firestore:"status_code" json:"statusCode"`
	LatencyMs  int       `firestore:"latency_ms" json:"latencyMs"`
}

type AppConfig struct {
	AutoApproveMode      string    `firestore:"auto_approve_mode" json:"autoApproveMode"`
	HackathonModeEnabled bool      `firestore:"hackathon_mode_enabled" json:"hackathonModeEnabled"`
	HackathonEndDate     string    `firestore:"hackathon_end_date" json:"hackathonEndDate"` // stored as ISO string
	InstanceType         string    `firestore:"instance_type" json:"instanceType"`
	KeysExpiresAtDate    time.Time `firestore:"keys_expires_at_date" json:"keysExpiresAtDate"`
	TrackStats           bool      `firestore:"track_stats" json:"trackStats"`
	UpdatedAt            time.Time `firestore:"updated_at" json:"updatedAt"`
	UpdatedBy            string    `firestore:"updated_by" json:"updatedBy"`
}

type InstanceState struct {
	InstanceID    string `firestore:"instance_id" json:"instanceId"`
	InstanceType  string `firestore:"instance_type" json:"instanceType"`
	State         string `firestore:"state" json:"state"` // "running" | "stopped"
	UptimeSeconds int64  `firestore:"uptime_seconds" json:"uptimeSeconds"`
	PublicIP      string `firestore:"public_ip" json:"publicIp"`
}
