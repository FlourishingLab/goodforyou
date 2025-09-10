package questions

import (
	"encoding/csv"
	"errors"
	"log"
	"os"

	"user-db/db"
	"user-db/shared"
)

// TODO make immutable? const?
var dimensions map[string]shared.Dimension
var dimensionQuestions []shared.Question
var questions map[int]shared.Question

var dimensionOrder map[string]int = map[string]int{
	"Physical Health":      1,
	"Mental Health":        2,
	"Social Relationships": 3,
	"Character & Virtue":   4,
	"Meaning & Purpose":    5,
	"Material Stability":   6,
	"Spirituality":         7,
}

func init() {
	err := loadQuestionsCSV()
	if err != nil {
		log.Fatalf("Failed to load questions: %v", err)
	}
}

func loadQuestionsCSV() error {
	questionsCSV, err := getFromFile()
	if err != nil {
		return err
	}

	reader := csv.NewReader(questionsCSV)
	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	questions = make(map[int]shared.Question)
	dimensions = make(map[string]shared.Dimension)

	for i, row := range rows {

		// TODO validate more
		if row[3] == "" {
			return errors.New("question text cannot be empty")
		}

		question := shared.Question{
			ID:           i,
			Dimension:    row[0],
			SubDimension: row[1],
			Facet:        row[2],
			Text:         row[3],
			MinLabel:     row[4],
			MaxLabel:     row[5],
		}
		questions[i] = question

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

func GetNextQuestions(userId string) ([]shared.Question, error) {

	// need answered questions
	userAnswer, exist := db.GetUser(userId)
	if !exist {
		return nil, errors.New("user not found")
	}

	// Are general dimension questions answered
	for _, v := range dimensionQuestions {
		if userAnswer.GetLatestAnswer(v.ID) == nil {
			return dimensionQuestions, nil
		}
	}

	// All general dimension questions answered, sort them
	sortedDimQ := userAnswer.SortByDimension(dimensionQuestions, dimensions)

	// start with the lowest Dimension
	for _, dim := range sortedDimQ {

		currentDimension := dimensions[dim.Name]

		// Are general subdimension questions answered?
		for _, v := range currentDimension.GeneralQuestions {
			if userAnswer.GetLatestAnswer(v.ID) == nil {
				return currentDimension.GeneralQuestions, nil
			}
		}

		// Send all questions from the first unanswered subdimension, if available
		subDim := currentDimension.SubDimensions
		for _, sd := range subDim {
			allAnswered := true
			subDimQs := []shared.Question{}
			for _, facets := range sd.Facets {
				for _, question := range facets.Questions {
					subDimQs = append(subDimQs, question)
					if userAnswer.GetLatestAnswer(question.ID) == nil {
						allAnswered = false
					}
				}
			}
			if !allAnswered {
				return subDimQs, nil
			}
		}
	}

	return []shared.Question{}, nil
}

func GetCompleteDimensions(ua db.UserAnswers) []string {

	// create copy of dimensions map
	dims := make(map[string]shared.Dimension)
	for k, v := range dimensions {
		//TODO temporarily remove dimensions without questions
		if k != "Happiness & Life Satisfaction" &&
			k != "Meaning & Purpose" &&
			k != "Character & Virtue" &&
			k != "Material Stability" &&
			k != "Spirituality" {
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

func getFromFile() (*os.File, error) {
	// for local testing only
	const FILE_PATH = "questions/questions.csv"
	return os.Open(FILE_PATH)
}

func GetDimensions() map[string]shared.Dimension {
	return dimensions
}

func GetQuestions() map[int]shared.Question {
	return questions
}
