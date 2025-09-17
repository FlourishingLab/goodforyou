package llm

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"user-db/shared"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

var client openai.Client
var ctx context.Context

func init() {
	ctx = context.Background()
	client = openai.NewClient()

}

func HolisticPrompt(sortedDimensions []shared.CatVal, sortedFacets []shared.CatVal) string {

	params := responses.ResponseNewParams{
		Prompt: responses.ResponsePromptParam{
			ID: "pmpt_68a854a4a3c48193ba6b74da1a8e866a0c7c540e5eb70354",
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.Opt[string]{Value: "Response in JSON\ndimension ratings:" + catValToString(sortedDimensions) + "\n\nfacets:\n" + catValToString(sortedFacets)},
		},
	}

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	sanitizedOutput, err := sanitizeAndExtractJSON(resp.OutputText())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Received Prompt for holistic")

	return sanitizedOutput
}

func DimensionPrompt(dimensionName string, dimensionRatings string) string {

	params := responses.ResponseNewParams{
		Prompt: responses.ResponsePromptParam{
			ID: "pmpt_68b6a4fd9d048196b3acf60938dc10040d196830d567e556",
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.Opt[string]{Value: "Response in JSON, Focus on Dimension " + dimensionName + "\nRatings:\n" + dimensionRatings},
		},
	}

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	sanitizedOutput, err := sanitizeAndExtractJSON(resp.OutputText())
	if err != nil {
		// TODO try again here once or twice?
		log.Fatal("json not valid", err)
	}

	log.Printf("Received Prompt for %s", dimensionName)

	return sanitizedOutput

}

func catValToString(sortedCat []shared.CatVal) (result string) {
	for _, cv := range sortedCat {
		result += cv.ToString() + "\n"
	}
	return result
}

// SanitizeAndExtractJSON attempts to find and extract a valid JSON object or array from a string.
// It's useful for cleaning up LLM responses that wrap JSON in markdown or other text.
func sanitizeAndExtractJSON(raw string) (string, error) {
	// First, check if the raw string is already valid JSON
	if json.Valid([]byte(raw)) {
		return raw, nil
	}

	// find the first {, neglect [ which would be valid JSON but not expected in this case
	startBrace := strings.Index(raw, "{")
	if startBrace == -1 {
		return "", errors.New("no Brace found in the string. Is it a valid JSON?")
	}

	// Find the last closing brace
	endBrace := strings.LastIndex(raw, "}")

	// Extract the potential JSON substring
	potentialJSON := raw[startBrace : endBrace+1]

	// Validate the extracted substring
	if json.Valid([]byte(potentialJSON)) {
		return potentialJSON, nil
	}

	return "", errors.New("extracted string is not valid JSON")
}
