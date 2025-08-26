package questions

import (
	"encoding/csv"
	"log"
	"os"
)

type Question struct {
	Dimension string `json:"dimension"`
	Facet     string `json:"facet"`
	Text      string `json:"text"`
	MinLabel  string `json:"minLabel"`
	MaxLabel  string `json:"maxLabel"`
}

type Facet struct {
	Name      string `json:"name"`
	Dimension string `json:"dimension"`
}

const SHEET_URL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTCQkU-Gxi1YHpT1qYAvXhbPNOd85CjBUMayXQUYUvMEJU3Yn8jkE1AveXrtAmJM8YHkyZRffZDegGk/pub?gid=0&single=true&output=csv"

var qs map[int]Question

var dimensions []string

var facets map[string]Facet

func init() {
	var err error
	qs, err = loadQuestionsCSV()
	if err != nil {
		log.Fatalf("Failed to load questions: %v", err)
	}

	facets = setFacets(qs)
	dimensions = setDimensions(qs)
}

func setFacets(qs map[int]Question) map[string]Facet {
	result := make(map[string]Facet)
	for _, q := range qs {
		result[q.Facet] = Facet{
			Name:      q.Facet,
			Dimension: q.Dimension,
		}
	}
	return result
}

func setDimensions(qs map[int]Question) (result []string) {
	seen := make(map[string]bool)
	for _, q := range qs {
		if !seen[q.Dimension] {
			seen[q.Dimension] = true
			result = append(result, q.Dimension)
		}
	}
	return result
}

func loadQuestionsCSV() (map[int]Question, error) {

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

	questions := make(map[int]Question)
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 5 {
			continue
		}
		questions[i] = Question{
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

func GetQuestions() map[int]Question {
	return qs
}
