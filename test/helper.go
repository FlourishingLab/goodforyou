package test

import (
	"user-db/db"
	"user-db/questions"
)

func AnswerSliceToAnswers(fakeAnswers map[int]int) map[int]db.QuestionAnswers {
	result := make(map[int]db.QuestionAnswers, len(fakeAnswers))
	for k, v := range fakeAnswers {
		result[k] = db.QuestionAnswers{
			LatestAnswer: db.AnswerEvent{
				Value: &v,
			},
		}
	}
	return result
}

func AllAnswers5And(setOtherValue []int) map[int]db.QuestionAnswers {

	result := make(map[int]db.QuestionAnswers)
	for i := 1; i <= len(questions.GetQuestions()); i++ {
		result[i] = db.QuestionAnswers{
			LatestAnswer: db.AnswerEvent{
				Value: func() *int {
					v := 5
					return &v
				}(),
			},
		}
	}
	for _, v := range setOtherValue {
		result[v] = db.QuestionAnswers{
			LatestAnswer: db.AnswerEvent{
				Value: func() *int {
					var val int
					if v == 12 {
						val = 1
					} else {
						val = 5
					}
					return &val
				}(),
			},
		}
	}
	return result
}
