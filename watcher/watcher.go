package watcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func Watch(bugsDir string) {
	running := false
	logrus.Infof("watcher started, watching %s", bugsDir)

	for {
		time.Sleep(2 * time.Second)

		if running {
			continue
		}

		files, err := filepath.Glob(filepath.Join(bugsDir, "*.md"))
		if err != nil || len(files) == 0 {
			continue
		}

		bugFile := files[0]
		logrus.Infof("found bug file %s, triggering claude", filepath.Base(bugFile))

		go func(path string) {
			running = true
			defer func() { running = false }()
			triggerClaude(path)
		}(bugFile)
	}
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

	cmd := exec.Command(claudePath, "--print", "--dangerously-skip-permissions", prompt)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("claude exited with error: %v", err)
		return
	}

	if err := os.Remove(bugFilePath); err != nil {
		logrus.Warnf("failed to delete bug file %s: %v", bugFilePath, err)
	} else {
		logrus.Infof("deleted %s after triggering claude", filepath.Base(bugFilePath))
	}
}
