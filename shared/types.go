package shared

import (
	"fmt"
)

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
	// Primary key in MongoDB
	ID        int `json:"_id"`
	Dimension string
	Facet     string
	Text      string
	MinLabel  string
	MaxLabel  string
}

type Facet struct {
	Name      string `json:"name"`
	Dimension string `json:"dimension"`
}

type AnswerKind int

const (
	DONTKNOW AnswerKind = iota
	SCALE
)

func (ct AnswerKind) String() string {
	switch ct {
	case DONTKNOW:
		return "dontknow"
	case SCALE:
		return "scale"
	default:
		return "Unknown"
	}
}
