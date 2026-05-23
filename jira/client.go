package jira

import (
	"encoding/json"
	"strings"
)

type WebhookPayload struct {
	WebhookEvent string `json:"webhookEvent"`
	Issue        Issue  `json:"issue"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary     string      `json:"summary"`
	Description Description `json:"description"`
	Priority    Named       `json:"priority"`
	Reporter    Reporter    `json:"reporter"`
	IssueType   Named       `json:"issuetype"`
	Status      Named       `json:"status"`
}

type Named    struct{ Name string `json:"name"` }
type Reporter struct{ DisplayName string `json:"displayName"` }

// Description handles both Jira v2 (plain string) and v3 (ADF object) formats.
type Description struct {
	plain string
	node  *ADFNode
}

func (d *Description) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		d.plain = s
		return nil
	}
	var node ADFNode
	if err := json.Unmarshal(b, &node); err != nil {
		return err
	}
	d.node = &node
	return nil
}

func (d *Description) PlainText() string {
	if d.plain != "" {
		return d.plain
	}
	if d.node != nil {
		return d.node.PlainText()
	}
	return ""
}

type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text"`
	Content []ADFNode `json:"content"`
}

func (n *ADFNode) PlainText() string {
	if n == nil {
		return ""
	}
	if n.Type == "text" {
		return n.Text
	}
	var b strings.Builder
	for _, child := range n.Content {
		b.WriteString(child.PlainText())
	}
	if n.Type == "paragraph" {
		b.WriteString("\n")
	}
	return b.String()
}

