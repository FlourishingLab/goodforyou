package questions_test

import (
	"testing"
	"user-db/db"
	"user-db/questions"
	"user-db/shared"
	"user-db/test"
)

func TestGetNextQuestions(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		userAnswers db.UserAnswers
		wantID      int
		wantErr     bool
	}{
		{
			name: "no-answers",
			userAnswers: db.UserAnswers{
				UserID:  "no-answers",
				Answers: map[int]db.QuestionAnswers{},
			},
			wantID:  3,
			wantErr: false,
		},
		{
			name: "spirituality-is-worse",
			userAnswers: db.UserAnswers{
				UserID: "spirituality-is-worse",
				Answers: test.AnswerSliceToAnswers(map[int]int{
					1:  5,
					2:  5,
					3:  5,
					4:  5,
					5:  5,
					6:  5,
					7:  5,
					8:  5,
					9:  5,
					10: 5,
					11: 5,
					12: 10,
					13: 1,
				}),
			},
			wantID:  45,
			wantErr: false,
		},
		{
			name: "material-stability-is-worst",
			userAnswers: db.UserAnswers{
				UserID: "material-stability-is-worst",
				Answers: test.AnswerSliceToAnswers(map[int]int{
					1:  5,
					2:  5,
					3:  5,
					4:  5,
					5:  5,
					6:  5,
					7:  5,
					8:  5,
					9:  5,
					10: 5,
					11: 5,
					12: 1,
					13: 3,
				}),
			},
			wantID:  49,
			wantErr: false,
		},
		{
			name: "material-answered-spirit-next",
			userAnswers: db.UserAnswers{
				UserID: "material-answered-spirit-next",
				Answers: test.AnswerSliceToAnswers(map[int]int{
					1:  5,
					2:  5,
					3:  5,
					4:  5,
					5:  5,
					6:  5,
					7:  5,
					8:  5,
					9:  5,
					10: 5,
					11: 5,
					12: 1, // material stability is the worst
					13: 3, // spirituality is the second worst
					46: 5,
					47: 5,
					48: 5,
					49: 5,
					50: 5,
				}),
			},
			wantID:  42, // Expecting spirituality question
			wantErr: false,
		},
		{
			name: "all-5",
			userAnswers: db.UserAnswers{
				UserID: "all-5",
				Answers: test.AnswerSliceToAnswers(map[int]int{
					1:  5,
					2:  5,
					3:  5,
					4:  5,
					5:  5,
					6:  5,
					7:  5,
					8:  5,
					9:  5,
					10: 5,
					11: 5,
					12: 5,
					13: 5,
				}),
			},
			wantID:  23,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := questions.GetNextQuestions(tt.userAnswers, 5)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetNextQuestions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetNextQuestions() succeeded unexpectedly")
			}
			if !qsWithID(got, tt.wantID) {
				t.Errorf("GetNextQuestions() = %v, expecting ID %v", got, tt.wantID)
			}
		})
	}
}

func qsWithID(got []shared.Question, wantID int) bool {
	for _, v := range got {
		if v.ID == wantID {
			return true
		}
	}
	return false
}
