package questions

import (
	"log"
	"sort"

	"user-db/shared"
)

type Answer struct {
	QuestionID int `json:"questionid"`
	Value      int `json:"value"`
}

func GetSorted(answers []Answer) (sortedDims []shared.CatVal, sortedFacets []shared.CatVal) {

	sortedDims = sortAvgCategory(answers, shared.DimensionType)
	sortedFacets = sortAvgCategory(answers, shared.FacetType)

	log.Printf("Dimensions: %v", sortedDims)
	log.Printf("Facets: %v", sortedFacets[:5])

	return sortedDims, sortedFacets
}

func sortAvgCategory(answers []Answer, catType shared.CatType) []shared.CatVal {

	mapC := make(map[string][]int)

	for _, answer := range answers {
		if question, exists := qs[answer.QuestionID]; exists {
			if catType == shared.DimensionType {
				mapC[question.Dimension] = append(mapC[question.Dimension], answer.Value)
			} else {
				mapC[question.Facet] = append(mapC[question.Facet], answer.Value)
			}
		}
	}

	var avgCategories []shared.CatVal
	for c, v := range mapC {
		avgCategories = append(avgCategories, shared.CatVal{CatType: catType, Name: c, Value: avg(v)})
	}

	// Sort by Value (ascending)
	sort.Slice(avgCategories, func(i, j int) bool {
		return avgCategories[i].Value < avgCategories[j].Value
	})

	return avgCategories
}

func avg(values []int) (total int) {
	for _, value := range values {
		total += value
	}
	return total / len(values)
}
