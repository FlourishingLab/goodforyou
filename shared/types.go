package shared

import "fmt"

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
