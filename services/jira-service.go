package services

import (
	"github.com/ThanvirXo/jira-auto-bug-solver/common"
	"github.com/ThanvirXo/jira-auto-bug-solver/jira"
	"github.com/ThanvirXo/jira-auto-bug-solver/writer"
	"github.com/sirupsen/logrus"
)

func (s *Service) HealthCheck() common.ResponseType {
	return common.SUCCESS
}

func (s *Service) HandleWebhook(payload *jira.WebhookPayload) error {
	if payload.Issue.Fields.IssueType.Name != "Bug" {
		logrus.Infof("ignoring non-bug issue: %s", payload.Issue.Fields.IssueType.Name)
		return nil
	}

	if err := writer.WriteBugMD(&payload.Issue); err != nil {
		logrus.Errorf("failed to write bug-context.md: %v", err)
		return err
	}

	logrus.Infof("bug-context.md written for %s", payload.Issue.Key)
	return nil
}
