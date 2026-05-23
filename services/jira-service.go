package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ThanvirXo/jira-auto-bug-solver/common"
	"github.com/ThanvirXo/jira-auto-bug-solver/jira"
	"github.com/ThanvirXo/jira-auto-bug-solver/writer"
	"github.com/sirupsen/logrus"
)

var bugQueue = make(chan string, 100)

func init() {
	go func() {
		for bugFilePath := range bugQueue {
			triggerClaude(bugFilePath)
		}
	}()
}

func (s *Service) HealthCheck() common.ResponseType {
	return common.SUCCESS
}

func (s *Service) HandleWebhook(payload *jira.WebhookPayload) error {
	if payload.Issue.Fields.IssueType.Name != "Bug" {
		logrus.Infof("ignoring non-bug issue: %s", payload.Issue.Fields.IssueType.Name)
		return nil
	}

	bugsDir := os.Getenv("FRONTEND_BUGS_DIR")
	bugFilePath := fmt.Sprintf("%s/%s.md", bugsDir, payload.Issue.Key)

	if err := writer.WriteBugMD(&payload.Issue); err != nil {
		logrus.Errorf("failed to write bug file: %v", err)
		return err
	}

	logrus.Infof("bug file written for %s, queued for claude", payload.Issue.Key)
	bugQueue <- bugFilePath

	return nil
}

func triggerClaude(bugFilePath string) {
	repoPath := os.Getenv("FRONTEND_REPO_PATH")
	if repoPath == "" {
		logrus.Error("FRONTEND_REPO_PATH not set")
		return
	}

	claudePath := os.Getenv("CLAUDE_PATH")
	prompt := `Read the file at ` + bugFilePath + `. Do not create a worktree for yourself. Spawn the frontend agent WITHOUT isolation worktree — the agent creates its own worktree internally, do not add another one on top.

The agent has strict rules:
- Fix ONLY the exact issue described in the bug. Nothing else.
- ONLY modify source code files that directly contain the bug.
- Do NOT touch any file in .claude/ directory under any circumstances. Not settings.json, not settings.local.json, nothing.
- Do NOT refactor, rename, reformat, or clean up anything outside the buggy code.
- Do NOT add new files unless the bug explicitly requires it.
- If unsure whether a file needs changing, do not change it.
- Do NOT push, do NOT commit, do NOT merge, do NOT switch branches. Stage the changes inside the worktree and stop. The user will review and merge manually.`

	logrus.Infof("triggering claude for %s", filepath.Base(bugFilePath))
	cmd := exec.Command(claudePath, "--print", "--dangerously-skip-permissions", prompt)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("claude exited with error: %v", err)
		return
	}

	if err := os.Remove(bugFilePath); err != nil {
		logrus.Warnf("failed to delete bug file: %v", err)
	} else {
		logrus.Infof("deleted %s after claude finished", filepath.Base(bugFilePath))
	}
}
