package paragraph

import (
	_ "embed"
	"encoding/json"
	"maps"
)

//go:embed paragraphs.json
var paragraphsJSON []byte

var paragraphs map[int]Paragraph

type Paragraph struct {
	Text  Text  `json:"text"`
	Links Links `json:"links"`
}

type Text struct {
	Title string `json:"title"`
	Short string `json:"short"`
	Long  string `json:"long"`
}

type Links struct {
	Inspirational string `json:"inspirational"`
}

func init() {
	json.Unmarshal(paragraphsJSON, &paragraphs)
}

func GetParagraphs() map[int]Paragraph {
	// Return a copy of the paragraphs map
	copy := make(map[int]Paragraph, len(paragraphs))
	maps.Copy(copy, paragraphs)
	return copy
}

func GetParagraph(i int) Paragraph {
	return paragraphs[i]
}
