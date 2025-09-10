package db

import (
	"encoding/json"
	"time"
)

type UserAnswers struct {
	// primary key in MongoDB
	UserID   string                  `json:"_id"`
	Answers  map[int]QuestionAnswers `json:"answers"`
	Insights map[string]Insight      `json:"insights"`
}

type QuestionAnswers struct {
	LatestAnswer AnswerEvent `json:"latestAnswer"`
	// History []AnswerEvent `json:"history" json:"history"`
}

type AnswerEvent struct {
	Kind      string    `json:"kind"`
	Value     *int      `json:"value,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Insight struct {
	InsightJson json.RawMessage `json:"insightJson"`
}
