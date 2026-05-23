package writer

import (
	"fmt"
	"os"
	"time"

	"github.com/ThanvirXo/jira-auto-bug-solver/jira"
)

func WriteBugMD(issue *jira.Issue) error {
	bugsDir := os.Getenv("FRONTEND_BUGS_DIR")

	if err := os.MkdirAll(bugsDir, 0755); err != nil {
		return err
	}

	backendPath := os.Getenv("BACKEND_REPO_PATH")
	frontendPath := os.Getenv("FRONTEND_REPO_PATH")
	outPath := fmt.Sprintf("%s/%s.md", bugsDir, issue.Key)

	content := fmt.Sprintf(`# Jira Bug: %s
> Received at %s

## Summary
%s

## Priority
%s

## Status
%s

## Reporter
%s

## Description
%s

## Repo Paths
- Backend: %s
- Frontend: %s

## Instructions for Claude
- Read this file completely before taking any action
- Based on the bug description, determine whether it belongs to the backend or frontend
- Navigate to the relevant repo path listed above
- Find the root cause by reading the relevant files
- Write the fix with the smallest possible diff
- Add or update tests if the fix touches testable logic
- Stage the changes only. Do not commit or open a PR
`,
		issue.Key,
		time.Now().Format("2006-01-02 15:04:05"),
		issue.Fields.Summary,
		issue.Fields.Priority.Name,
		issue.Fields.Status.Name,
		issue.Fields.Reporter.DisplayName,
		issue.Fields.Description.PlainText(),
		backendPath,
		frontendPath,
	)

	return os.WriteFile(outPath, []byte(content), 0644)
}
