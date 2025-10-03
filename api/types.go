package api

import (
	"user-db/paragraph"
	"user-db/shared"
)

type ResponsePayload struct {
	Answers []HttpAnswer `json:"answers"`
}

type HttpAnswer struct {
	QuestionID int    `json:"questionid"`
	Value      int    `json:"value"`
	Kind       string `json:"kind"`
}

type contextKey string

type ConsistentlyResponse struct {
	Streak    int                 `json:"streak"`
	Paragraph paragraph.Paragraph `json:"paragraph"`
	Questions []shared.Question   `json:"questions"`
}
