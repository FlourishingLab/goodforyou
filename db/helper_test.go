package db_test

import (
	"fmt"
	"strings"
	"testing"
	"user-db/db"
	"user-db/questions"
	"user-db/shared"
	"user-db/test"
)

func TestUserAnswers_GetSorted(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		userID string
		// Named input parameters for target function.
		userAnswers db.UserAnswers
		qs          map[int]shared.Question
		dims        map[string]shared.Dimension
		want        []string
		want2       []string
	}{
		{
			name:   "all-answers",
			userID: "all-answers",
			userAnswers: db.UserAnswers{
				UserID:  "all-answers-5",
				Answers: test.AllAnswers5And([]int{}),
			},
			qs:   questions.GetQuestions(),
			dims: questions.GetDimensions(),
			want: []string{"Physical Health", "Mental Health", "Social Relationships", "Character & Virtue", "Meaning & Purpose", "Spirituality", "Material Stability", "Happiness & Life Satisfaction"},
			// Fails cause it is non deterministic
			want2: []string{"Boundaries.Assertive Limit-Setting",
				"Financial Planning.Saving & Investing",
				"Awe & Transcendence.Connection ",
				"Boundaries.Enforcing Boundaries",
				"Values & Authenticity.Values–Action Congruence",
				"Awe & Transcendence.Contemplation",
				"Sleep.circadian rhythm",
				"Emotion Regulation.Awareness & Labeling",
				"Sleep.Sleep quality",
				"Cognitive Control.Sustained Attention",
				"Sleep.alertness",
				"Activity.Aerobic",
				"Boundaries.Personal Autonomy",
				"Communication.Active Listening and Empathy",
				"Financial Planning.Debt Management",
				"Connection.Social Integration",
				"Activity.Strength",
				"Cognitive Control.Goal Maintenance",
				"Values & Authenticity.Courageous Authenticity",
				"Financial Planning.Cashflow Plan & Tracking",
				"Awe & Transcendence.Guiding Beliefs",
				"Emotion Regulation.Acceptance",
				"Financial Planning.Liquidity",
				"Emotion Regulation.Reappraisal ",
				"Communication.Open and Honest Expression",
				"Connection.Belonging",
				"Values & Authenticity.Values Clarity",
				"Cognitive Control.Inhibitory Control",
				"Values & Authenticity.Identity Coherence",
				"Boundaries.Emotional Boundaries",
				"Connection.Emotional Support",
				"Financial Planning.Payments Reliability",
				"Activity.Sedentary Behaviour",
				"Awe & Transcendence.Wonder"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got2 := tt.userAnswers.GetSorted(tt.qs, tt.dims)
			if err := compareDimensionToCatVal(got, tt.want); err != nil {
				t.Errorf("GetSorted() Dimensions = %v", err)
			}
			if err := compareFacetsToCatVal(got2, tt.want2); err != nil {
				t.Errorf("GetSorted() Facets = %v", err)
			}
		})
	}
}

func compareDimensionToCatVal(got []shared.CatVal, want []string) error {
	for i, v := range got {
		// get rid of Dimension prefix
		nameWOPrefix, _ := strings.CutPrefix(v.Name, "Dimension ")
		if nameWOPrefix != want[i] {
			return fmt.Errorf("%s unequal to %s", nameWOPrefix, want[i])
		}
	}
	return nil
}
func compareFacetsToCatVal(got []shared.CatVal, want []string) error {
	for i, v := range got {

		nameWOPrefix, _ := strings.CutPrefix(v.Name, "Facet ")
		if nameWOPrefix != want[i] {
			return fmt.Errorf("%s unequal to %s", nameWOPrefix, want[i])
		}
	}
	return nil
}
