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

func (ua *UserAnswers) SortByDimension(qs []shared.Question, dims map[string]shared.Dimension) []shared.CatVal {

	dimsToQuestions := make(map[string][]int)

	for _, question := range qs {
		answer := ua.GetLatestAnswer(question.ID)
		if answer != nil {
			dimsToQuestions[question.Dimension] = append(dimsToQuestions[question.Dimension], *answer.Value)
		}
	}

	var sortedDimensions []shared.CatVal
	for dimName, v := range dimsToQuestions {
		sortedDimensions = append(sortedDimensions, shared.CatVal{CatType: shared.DimensionType, Name: dimName, Value: avg(v)})
	}

	// Sort by Value (ascending)
	sort.Slice(sortedDimensions, func(i, j int) bool {

		if sortedDimensions[i].Value != sortedDimensions[j].Value {
			return sortedDimensions[i].Value < sortedDimensions[j].Value
		}
		// if equal rating, prioritise specific dimensions
		return dims[sortedDimensions[i].Name].Rank < dims[sortedDimensions[j].Name].Rank
	})

	return sortedDimensions
}

func (ua *UserAnswers) SortByFacet(qs []shared.Question) []shared.CatVal {

	facetsToQuestions := make(map[string][]int)

	for _, question := range qs {
		if question.Facet != shared.GENERAL {
			answer := ua.GetLatestAnswer(question.ID)
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
					qs := []int{}
					for _, q := range facet.Questions {
						qs = append(qs, *ua.GetLatestAnswer(q.ID).Value)
					}
					facetAvg := avg(qs)
					result += fk + ": " + strconv.Itoa(facetAvg)
				}
			}
		}
	}
	return result

}

func (ua *UserAnswers) HasInsight(insightName string) bool {
	insight, ok := ua.Insights[insightName]
	if ok && insight.Status == DONE {
		return true
	}
	return false
}

func (ua *UserAnswers) NeedsInsight(insightName string) bool {
	_, ok := ua.Insights[insightName]
	return !ok
}

func (ua *UserAnswers) GetInsight(insightName string) json.RawMessage {
	v := ua.Insights[insightName]
	return v.InsightJson
}

func avg(values []int) (total int) {
	for _, value := range values {
		total += value
	}
	return total / len(values)
}

func (ua *UserAnswers) GetSorted(qs map[int]shared.Question, dims map[string]shared.Dimension) (sortedDims []shared.CatVal, sortedFacets []shared.CatVal) {

	// convert questions to list of questions
	questionsList := shared.MapToSlice(qs)

	sortedDims = ua.SortByDimension(questionsList, dims)
	sortedFacets = ua.SortByFacet(questionsList)

	log.Printf("Dimensions: %v", sortedDims)
	log.Printf("Facets: %v", sortedFacets[:5])

	return sortedDims, sortedFacets
}

func (ua *UserAnswers) IDOfNextParagraph() int {
	highest := 0
	for k, v := range ua.Paragraphs {
		if v.WasShown {
			if k > highest {
				highest = k
			}
		}
	}
	return highest + 1
}
