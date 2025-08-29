package questions

import (
	"encoding/csv"
	"log"
	"os"
	"user-db/shared"
)

// const SHEET_URL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTCQkU-Gxi1YHpT1qYAvXhbPNOd85CjBUMayXQUYUvMEJU3Yn8jkE1AveXrtAmJM8YHkyZRffZDegGk/pub?gid=0&single=true&output=csv"

var qs map[int]shared.Question

// var dimensions []string

// var facets map[string]shared.Facet

func init() {
	var err error
	qs, err = loadQuestionsCSV()
	if err != nil {
		log.Fatalf("Failed to load questions: %v", err)
	}

	// facets = setFacets(qs)
	// dimensions = setDimensions(qs)
}

// func setFacets(qs map[int]shared.Question) map[string]shared.Facet {
// 	result := make(map[string]shared.Facet)
// 	for _, q := range qs {
// 		result[q.Facet] = shared.Facet{
// 			Name:      q.Facet,
// 			Dimension: q.Dimension,
// 		}
// 	}
// 	return result
// }

// func setDimensions(qs map[int]shared.Question) (result []string) {
// 	seen := make(map[string]bool)
// 	for _, q := range qs {
// 		if !seen[q.Dimension] {
// 			seen[q.Dimension] = true
// 			result = append(result, q.Dimension)
// 		}
// 	}
// 	return result
// }

func loadQuestionsCSV() (map[int]shared.Question, error) {

	// questionsCSV, err := getFromGoogleSheet()
	questionsCSV, err := getFromFile()
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(questionsCSV)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	questions := make(map[int]shared.Question)
	for i, row := range rows {
		questions[i] = shared.Question{
			ID:        i,
			Dimension: row[0],
			Facet:     row[1],
			Text:      row[2],
			MinLabel:  row[3],
			MaxLabel:  row[4],
		}
	}

	return questions, nil
}

func getFromFile() (*os.File, error) {
	// for local testing only
	const FILE_PATH = "questions/questions.csv"
	return os.Open(FILE_PATH)
}

// func getFromGoogleSheet(url string) (io.ReadCloser, error) {
// 	resp, err := http.Get(SHEET_URL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed to fetch CSV: status code %d", resp.StatusCode)
// 	}
// 	return resp.Body, nil
// }

func GetQuestions() map[int]shared.Question {
	return qs
}
