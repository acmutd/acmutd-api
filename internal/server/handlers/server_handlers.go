package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// GetInstanceState returns the current instance state from app_config.
func (h *Handler) GetInstanceState(c *gin.Context) {
	state, err := h.db.GetInstanceState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch instance state"})
		return
	}
	c.JSON(http.StatusOK, state)
}

// StartServer records a "start" server state log and sets instance_state to "running". Admin only.
// Actual EC2 control is out of scope — this records intent and persists the state change.
func (h *Handler) StartServer(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	state, err := h.db.GetInstanceState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch instance state"})
		return
	}

	newState := *state
	newState.State = "running"
	newState.UptimeSeconds = 0

	if err := h.db.UpdateInstanceState(c.Request.Context(), newState); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update instance state"})
		return
	}

	h.recordServerStateLog(c, uid, "start", state.State, "running", state.InstanceType, state.InstanceType, "Manual start from admin dashboard")
	c.JSON(http.StatusOK, gin.H{"message": "server start recorded"})
}

// StopServer records a "stop" server state log and sets instance_state to "stopped". Admin only.
func (h *Handler) StopServer(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	state, err := h.db.GetInstanceState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch instance state"})
		return
	}

	newState := *state
	newState.State = "stopped"

	if err := h.db.UpdateInstanceState(c.Request.Context(), newState); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update instance state"})
		return
	}

	h.recordServerStateLog(c, uid, "stop", state.State, "stopped", state.InstanceType, state.InstanceType, "Manual stop from admin dashboard")
	c.JSON(http.StatusOK, gin.H{"message": "server stop recorded"})
}

// EnableHackathonMode scales instance type to t3.large and enables hackathon mode. Admin only.
func (h *Handler) EnableHackathonMode(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	state, err := h.db.GetInstanceState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch instance state"})
		return
	}

	const hackathonType = "t3.large"
	newState := *state
	newState.InstanceType = hackathonType

	if err := h.db.UpdateInstanceState(c.Request.Context(), newState); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update instance state"})
		return
	}

	cfgUpdates := map[string]any{
		"hackathon_mode_enabled": true,
		"instance_type":          hackathonType,
	}
	if _, err := h.db.UpdateAppConfig(c.Request.Context(), cfgUpdates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	h.recordServerStateLog(c, uid, "hackathon_enable", state.State, state.State, state.InstanceType, hackathonType, "Hackathon mode enabled from admin dashboard")
	c.JSON(http.StatusOK, gin.H{"message": "hackathon mode enabled"})
}

// DisableHackathonMode scales instance type back to t3.micro and disables hackathon mode. Admin only.
func (h *Handler) DisableHackathonMode(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	state, err := h.db.GetInstanceState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch instance state"})
		return
	}

	const defaultType = "t3.micro"
	newState := *state
	newState.InstanceType = defaultType

	if err := h.db.UpdateInstanceState(c.Request.Context(), newState); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update instance state"})
		return
	}

	cfgUpdates := map[string]any{
		"hackathon_mode_enabled": false,
		"instance_type":          defaultType,
	}
	if _, err := h.db.UpdateAppConfig(c.Request.Context(), cfgUpdates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	h.recordServerStateLog(c, uid, "hackathon_disable", state.State, state.State, state.InstanceType, defaultType, "Hackathon mode disabled from admin dashboard")
	c.JSON(http.StatusOK, gin.H{"message": "hackathon mode disabled"})
}

func (h *Handler) recordServerStateLog(c *gin.Context, actorUID, action, prevState, newState, prevType, newType, reason string) {
	user, _ := h.db.GetUser(c.Request.Context(), actorUID)
	actorEmail := ""
	if user != nil {
		actorEmail = user.Email
	}

	b := make([]byte, 6)
	rand.Read(b)
	logID := "srv_" + hex.EncodeToString(b)

	b2 := make([]byte, 6)
	rand.Read(b2)
	reqID := "req_" + hex.EncodeToString(b2)

	log := types.ServerStateLog{
		LogID:         logID,
		Timestamp:     time.Now(),
		Action:        action,
		Status:        "success",
		TriggerSource: "admin_dashboard",
		ActorType:     "user",
		ActorID:       actorUID,
		ActorEmail:    actorEmail,
		Reason:        reason,
		RequestID:     reqID,
		PreviousState: prevState,
		NewState:      newState,
		PreviousType:  prevType,
		NewType:       newType,
	}

	// Fire-and-forget; log failure is non-critical
	go h.db.CreateServerStateLog(c.Request.Context(), log)
}
