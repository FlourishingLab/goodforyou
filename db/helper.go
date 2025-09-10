package db

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"user-db/shared"
)

func (ua *UserAnswers) GetLatestAnswer(questionID int) *AnswerEvent {

	if qa, ok := ua.Answers[questionID]; ok {
		return &qa.LatestAnswer
	}
	return nil
}

func (ua *UserAnswers) SortByDimension(questions []shared.Question) []shared.CatVal {

	dimsToQuestions := make(map[string][]int)

	for i, question := range questions {
		answer := ua.GetLatestAnswer(i)
		if answer != nil {
			dimsToQuestions[question.Dimension] = append(dimsToQuestions[question.Dimension], *answer.Value)
		}

	}

	var sortedDimensions []shared.CatVal
	for i, v := range dimsToQuestions {
		sortedDimensions = append(sortedDimensions, shared.CatVal{CatType: shared.DimensionType, Name: i, Value: avg(v)})
	}

	// Sort by Value (ascending)
	sort.Slice(sortedDimensions, func(i, j int) bool {
		return sortedDimensions[i].Value < sortedDimensions[j].Value
	})

	return sortedDimensions
}

func (ua *UserAnswers) SortByFacet(questions []shared.Question) []shared.CatVal {

	facetsToQuestions := make(map[string][]int)

	for i, question := range questions {
		if question.Facet != shared.GENERAL {
			answer := ua.GetLatestAnswer(i)
			if answer != nil {
				facetsToQuestions[question.SubDimension+"."+question.Facet] = append(facetsToQuestions[question.Facet], *answer.Value)
			}
		}
	}

	var sortedFacets []shared.CatVal
	for subDimDotFacet, v := range facetsToQuestions {
		sortedFacets = append(sortedFacets, shared.CatVal{CatType: shared.FacetType, Name: subDimDotFacet, Value: avg(v)})
	}

	// Sort by Value (ascending)
	sort.Slice(sortedFacets, func(i, j int) bool {
		return sortedFacets[i].Value < sortedFacets[j].Value
	})

	return sortedFacets
}

func (ua *UserAnswers) DimensionRatingsToString(dimensionName string, dimensions map[string]shared.Dimension) string {

	var result string

	for dk, dims := range dimensions {
		if dk == dimensionName {
			for sk, subdims := range dims.SubDimensions {
				result += sk + ":\n"
				for fk, facet := range subdims.Facets {
					questions := []int{}
					for _, q := range facet.Questions {
						questions = append(questions, *ua.GetLatestAnswer(q.ID).Value)
					}
					facetAvg := avg(questions)
					result += fk + ": " + strconv.Itoa(facetAvg)
				}
			}
		}
	}
	return result

}

func (ua *UserAnswers) HasInsight(insightName string) bool {
	_, ok := ua.Insights[insightName]
	return ok
}

func (ua *UserAnswers) GetInsight(insightName string) json.RawMessage {
	v, _ := ua.Insights[insightName]
	return v.InsightJson
}

func avg(values []int) (total int) {
	for _, value := range values {
		total += value
	}
	return total / len(values)
}

func (ua *UserAnswers) GetSorted(questions map[int]shared.Question) (sortedDims []shared.CatVal, sortedFacets []shared.CatVal) {

	// convert questions to list of questions
	questionsList := shared.MapToSlice(questions)

	sortedDims = ua.SortByDimension(questionsList)
	sortedFacets = ua.SortByFacet(questionsList)

	log.Printf("Dimensions: %v", sortedDims)
	log.Printf("Facets: %v", sortedFacets[:5])

	return sortedDims, sortedFacets
}
