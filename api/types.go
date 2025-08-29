package api

import (
	"user-db/shared"
)

type ResponsePayload struct {
	UserID  string       `json:"userId"`
	Answers []HttpAnswer `json:"answers"`
}

type HttpAnswer struct {
	QuestionID int               `json:"questionid"`
	Value      int               `json:"value"`
	Kind       shared.AnswerKind `json:"kind"`
}
