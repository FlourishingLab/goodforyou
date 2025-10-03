package questions

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"log"
	"maps"
	"strings"

	"user-db/db"
	"user-db/shared"
)

// Embed the questions.csv file
//
//go:embed questions.csv
var questionsCSV []byte

var dimensions map[string]shared.Dimension
var dimensionQuestions []shared.Question
var questions map[int]shared.Question

var dimensionOrder map[string]int = map[string]int{
	"Physical Health":      1,
	"Mental Health":        2,
	"Social Relationships": 3,
	"Character & Virtue":   4,
	"Meaning & Purpose":    5,
	"Spirituality":         6,
	"Material Stability":   7,
}

func init() {
	err := loadQuestionsCSV()
	if err != nil {
		log.Fatalf("Failed to load questions: %v", err)
	}
}

func loadQuestionsCSV() error {

	// Use the embedded questionsCSV data
	reader := csv.NewReader(strings.NewReader(string(questionsCSV)))
	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	questions = make(map[int]shared.Question)
	dimensions = make(map[string]shared.Dimension)

	for i, row := range rows {

		// start with 1 to align with the google sheet
		questionNumber := i + 1
		// TODO validate more
		if row[3] == "" {
			return errors.New("question text cannot be empty")
		}

		question := shared.Question{
			ID:           questionNumber,
			Dimension:    row[0],
			SubDimension: row[1],
			Facet:        row[2],
			Text:         row[3],
			MinLabel:     row[4],
			MaxLabel:     row[5],
		}
		questions[questionNumber] = question

		// init dimensions
		dimension, ok := dimensions[question.Dimension]
		if !ok {
			rank, ok := dimensionOrder[question.Dimension]
			if !ok {
				rank = 100
			}

			dimension = shared.Dimension{
				SubDimensions:    make(map[string]shared.SubDimension),
				GeneralQuestions: []shared.Question{},
				Rank:             rank,
			}
		}

		if question.SubDimension == shared.GENERAL {
			dimensionQuestions = append(dimensionQuestions, question)
		} else if question.Facet == shared.GENERAL {
			dimension.GeneralQuestions = append(dimension.GeneralQuestions, question)
		} else {
			subDim, ok := dimension.SubDimensions[question.SubDimension]
			if !ok {
				subDim = shared.SubDimension{
					Facets: make(map[string]shared.Facet),
				}
			}
			// add non-general to facets
			facet, ok := subDim.Facets[question.Facet]
			if !ok {
				facet = shared.Facet{
					Questions: []shared.Question{},
				}
			}
			facet.Questions = append(facet.Questions, question)
			subDim.Facets[question.Facet] = facet
			dimension.SubDimensions[question.SubDimension] = subDim

		}

		dimensions[question.Dimension] = dimension
	}
	return nil
}

func GetNextQuestions(userAnswers db.UserAnswers, max int) ([]shared.Question, error) {

	unansweredQuestions := []shared.Question{}

	// Are general dimension questions answered
	for _, v := range dimensionQuestions {
		if userAnswers.GetLatestAnswer(v.ID) == nil {
			unansweredQuestions = append(unansweredQuestions, v)
		}
	}

	if len(unansweredQuestions) >= max {
		return unansweredQuestions[:max], nil
	}

	// All general dimension questions answered, sort them
	sortedDimQ := userAnswers.SortByDimension(dimensionQuestions, GetDimensions())

	// start with the lowest Dimension
	for _, dim := range sortedDimQ {

		currentDimension := GetDimensions()[dim.Name]

		// Are general subdimension questions answered?
		for _, v := range currentDimension.GeneralQuestions {
			if userAnswers.GetLatestAnswer(v.ID) == nil {
				unansweredQuestions = append(unansweredQuestions, v)
			}
		}

		// Send all questions from the first unanswered subdimension, if available
		subDim := currentDimension.SubDimensions
		for _, sd := range subDim {
			for _, facets := range sd.Facets {
				for _, question := range facets.Questions {
					if userAnswers.GetLatestAnswer(question.ID) == nil {
						unansweredQuestions = append(unansweredQuestions, question)
					}
				}
			}
			if len(unansweredQuestions) >= max {
				return unansweredQuestions[:max], nil
			}
		}
	}

	return unansweredQuestions, nil
}

func GetCompleteDimensions(ua db.UserAnswers) []string {

	// create copy of dimensions map
	dims := make(map[string]shared.Dimension)
	for k, v := range dimensions {
		//TODO temporarily remove dimensions without questions
		if k != "Happiness & Life Satisfaction" &&
			k != "Character & Virtue" {
			dims[k] = v
		}
	}

	for i, v := range questions {
		if ua.GetLatestAnswer(i) == nil {
			delete(dims, v.Dimension)
		}
	}

	return shared.GetKeysFromMap(dims)
}

func GetDimensions() map[string]shared.Dimension {
	// Return a copy of the dimensions map
	copy := make(map[string]shared.Dimension, len(dimensions))
	maps.Copy(copy, dimensions)
	return copy
}

func GetQuestions() map[int]shared.Question {
	// Return a copy of the questions map
	copy := make(map[int]shared.Question, len(questions))
	maps.Copy(copy, questions)
	return copy
}

func GetDimensionQuestions() []shared.Question {
	// Return a copy of the questions map
	copy := make([]shared.Question, len(questions))
	return copy
}
