package questions

import (
	"log"

	"user-db/db"
	"user-db/shared"
)

func GetSorted(userAnswers db.UserAnswers) (sortedDims []shared.CatVal, sortedFacets []shared.CatVal) {

	// convert questions to list of questions
	questionsList := mapToSlice(questions)

	sortedDims = userAnswers.SortByDimension(questionsList)
	sortedFacets = userAnswers.SortByFacet(questionsList)

	log.Printf("Dimensions: %v", sortedDims)
	log.Printf("Facets: %v", sortedFacets[:5])

	return sortedDims, sortedFacets
}

// func sortAvgCategory(userAnswers db.UserAnswers, catType shared.CatType) []shared.CatVal {

// 	mapC := make(map[string][]int)

// 	answers := userAnswers.Answers

// 	for i, answer := range answers {
// 		if question, exists := questions[i]; exists {
// 			if catType == shared.DimensionType {
// 				mapC[question.Dimension] = append(mapC[question.Dimension], *answer.LatestAnswer.Value)
// 			} else {
// 				mapC[question.Facet] = append(mapC[question.Facet], *answer.LatestAnswer.Value)
// 			}
// 		}
// 	}

// 	var avgCategories []shared.CatVal
// 	for c, v := range mapC {
// 		avgCategories = append(avgCategories, shared.CatVal{CatType: catType, Name: c, Value: avg(v)})
// 	}

// 	// Sort by Value (ascending)
// 	sort.Slice(avgCategories, func(i, j int) bool {
// 		return avgCategories[i].Value < avgCategories[j].Value
// 	})

// 	return avgCategories
// }

// func avg(values []int) (total int) {
// 	for _, value := range values {
// 		total += value
// 	}
// 	return total / len(values)
// }
