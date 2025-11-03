package shared

import (
	"fmt"
)

const GENERAL string = "general"
const HABITS string = "Habits"

type CatType int

const (
	DimensionType CatType = iota
	FacetType
)

func (ct CatType) String() string {
	switch ct {
	case DimensionType:
		return "Dimension"
	case FacetType:
		return "Facet"
	default:
		return "Unknown"
	}
}

type CatVal struct {
	CatType CatType // Dimension or Facet
	Name    string
	Value   int
}

func (cv CatVal) ToString() string {
	return fmt.Sprintf("%s: %d", cv.Name, cv.Value)
}

type Question struct {
	ID           int    `json:"id"`
	Text         string `json:"text"`
	MinLabel     string `json:"min_label"`
	MaxLabel     string `json:"max_label"`
	Dimension    string `json:"dimension"`
	SubDimension string `json:"sub_dimension"`
	Facet        string `json:"facet"`
}

type Facet struct {
	Questions []Question `json:"questions"`
}

type Dimension struct {
	SubDimensions    map[string]SubDimension `json:"sub_dimensions"`
	GeneralQuestions []Question              `json:"general_questions"`
	Rank             int                     `json:"rank,omitempty"`
}

type SubDimension struct {
	Facets map[string]Facet `json:"facets"`
}

type AnswerKind string

const (
	DONTKNOW AnswerKind = "DONTKNOW"
	SCALE    AnswerKind = "SCALE"
)

func (ak AnswerKind) String() string {
	switch ak {
	case DONTKNOW:
		return "DONTKNOW"
	case SCALE:
		return "SCALE"
	default:
		return "UNKNOWN"
	}
}

func ToAnswerKind(kind string) (AnswerKind, error) {
	switch kind {
	case "DONTKNOW":
		return DONTKNOW, nil
	case "SCALE":
		return SCALE, nil
	}
	return AnswerKind(""), fmt.Errorf("invalid AnswerKind: %s", kind)
}
