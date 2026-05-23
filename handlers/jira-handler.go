package handlers

import (
	"encoding/json"
	"io"

	"github.com/ThanvirXo/jira-auto-bug-solver/common"
	"github.com/ThanvirXo/jira-auto-bug-solver/jira"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) HealthCheck(c *gin.Context) {
	status := h.Services.HealthCheck()
	common.NewResponse(status, "Health check successful").Respond(c)
}

func (h *Handler) JiraWebhook(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	logrus.Infof("jira webhook raw body: %s", string(body))

	var payload jira.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logrus.Errorf("failed to parse webhook payload: %v", err)
		common.NewResponse(common.ERROR, "Invalid payload").Respond(c)
		return
	}

	if err := h.Services.HandleWebhook(&payload); err != nil {
		common.NewResponse(common.SERVER_ERROR, "Failed to process webhook").Respond(c)
		return
	}

	common.NewResponse(common.SUCCESS, "ok").Respond(c)
}
