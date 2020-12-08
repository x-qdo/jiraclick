package clickup

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type EventType string

const (
	TaskCreated             EventType = "taskCreated"
	TaskUpdated             EventType = "taskUpdated"
	TaskDeleted             EventType = "taskDeleted"
	TaskPriorityUpdated     EventType = "taskPriorityUpdated"
	TaskStatusUpdated       EventType = "taskStatusUpdated"
	TaskAssigneeUpdated     EventType = "taskAssigneeUpdated"
	TaskDueDateUpdated      EventType = "taskDueDateUpdated"
	TaskTagUpdated          EventType = "taskTagUpdated"
	TaskMoved               EventType = "taskMoved"
	TaskCommentPosted       EventType = "taskCommentPosted"
	TaskCommentUpdated      EventType = "taskCommentUpdated"
	TaskTimeEstimateUpdated EventType = "taskTimeEstimateUpdated"
	TaskTimeTrackedUpdated  EventType = "taskTimeTrackedUpdated"
)

type WebhookEvent struct {
	ID      string        `json:"webhook_id"`
	Type    EventType     `json:"event"`
	Changes []HistoryItem `json:"history_items"`
	TaskID  string        `json:"task_id"`
}

type HistoryItem struct {
	ID     string      `json:"id"`
	Type   int         `json:"type"`
	Date   string      `json:"date"`
	Field  string      `json:"field"`
	User   User        `json:"user"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

func (item *HistoryItem) CovertTo() {

}

func CheckSignature(signature, body, secret string) bool {
	s := []byte(secret)
	m := []byte(body)

	hash := hmac.New(sha256.New, s)
	hash.Write(m)

	// to lowercase hexits
	result := hex.EncodeToString(hash.Sum(nil))

	return result == signature
}

func ParseEvent(body string) (*WebhookEvent, error) {
	e := new(WebhookEvent)
	if err := json.Unmarshal([]byte(body), e); err != nil {
		return nil, err
	}
	if len(e.Changes) == 0 {
		return nil, nil
	}

	return e, nil
}
