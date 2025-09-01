package llm

import (
	"context"
	"log"
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

func Prompt(sortedDimensions []shared.CatVal, sortedFacets []shared.CatVal) string {

	params := responses.ResponseNewParams{
		Prompt: responses.ResponsePromptParam{
			ID: "pmpt_68a854a4a3c48193ba6b74da1a8e866a0c7c540e5eb70354",
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.Opt[string]{Value: "Response in JSON\ncategories:\n" + catValToString(sortedDimensions) + "\nweakest subcategories\n:" + catValToString(sortedFacets)},
		},
	}

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.OutputText())

	return resp.OutputText()

}

func catValToString(sortedCat []shared.CatVal) (result string) {
	for _, cv := range sortedCat {
		result += cv.ToString() + "\n"
	}
	return result
}
