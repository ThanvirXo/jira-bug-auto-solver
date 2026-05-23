package classifier

import (
	"context"
	"fmt"
	"os"
	"strings"

	einoai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/sirupsen/logrus"
)

var chatModel *einoai.ChatModel

func Init() {
	ctx := context.Background()
	baseURL := os.Getenv("LLM_URL")

	var err error
	chatModel, err = einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		APIKey:  os.Getenv("VLLM_KEY"),
		BaseURL: baseURL,
		Model:   os.Getenv("LLM_MODEL"),
	})
	if err != nil {
		logrus.Fatalf("failed to create classifier model: %v", err)
	}
}

const systemPrompt = `You are a senior engineer doing bug triage. Your job is to determine whether a bug needs to be fixed in the FRONTEND codebase or the BACKEND codebase.

CRITICAL CONTEXT: Bug reporters are often non-technical — product managers, QA, or end users. They describe symptoms, not causes. You must reason about the root cause and where a developer would need to make a code change to fix it.

DECISION FRAMEWORK — reason through these internally before answering:

1. DATA NOT SHOWING / BLANK / "NOTHING IS COMING" (most ambiguous case)
   → "not being received", "nothing comes back", "response is empty", "API returns nothing" → BACKEND
   → "data loads but doesn't appear", "table blank even though API works", "state set but UI doesn't update" → FRONTEND
   → Reporter just says "nothing shows up" / "not loading" with no technical context → lean BACKEND by default

2. VISUAL / LAYOUT SYMPTOMS
   → Broken layout, wrong colour, misaligned, CSS, animation broken → FRONTEND
   → Button doesn't respond, form won't submit, modal won't open, dropdown empty → FRONTEND
   → Page crashes with JS error → FRONTEND

3. DATA INTEGRITY / CORRECTNESS
   → Wrong value saved, not persisted, lost after refresh → BACKEND
   → Business logic wrong on server → BACKEND
   → Correct data from API but displayed/formatted wrong on screen → FRONTEND

4. PERFORMANCE / AVAILABILITY
   → Timeout, 500/503 error, slow response, server down → BACKEND

5. KEYWORD SIGNALS (tiebreakers only)
   BACKEND: API, endpoint, server, database, 500, 404, timeout, payload, null data, not saved, queue, cron
   FRONTEND: button, page, component, form, table, chart, dropdown, modal, CSS, layout, not visible, blank, UI, render, React, Vue

IMPORTANT: Reason through the framework internally. Output ONLY a single line — no explanation, no reasoning, no other text:
CLASSIFICATION: frontend
or
CLASSIFICATION: backend`

const userPrompt = `BUG REPORT:
Summary: %s
Description: %s`

func IsFrontend(summary, description string) (bool, error) {
	msg, err := chatModel.Generate(context.Background(), []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: fmt.Sprintf(userPrompt, summary, description)},
	})
	if err != nil {
		return false, err
	}

	logrus.Infof("classifier response: %s", msg.Content)

	for _, line := range strings.Split(msg.Content, "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(line, "classification:") {
			label := strings.TrimSpace(strings.TrimPrefix(line, "classification:"))
			return label == "frontend", nil
		}
	}

	logrus.Warnf("classifier response missing CLASSIFICATION line, falling back to keyword scan")
	return strings.Contains(strings.ToLower(msg.Content), "frontend"), nil
}
