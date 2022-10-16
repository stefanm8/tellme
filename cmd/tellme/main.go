package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v2"
)

type Choice struct {
	Text string `json:"text"`
}
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

var openaiURI = "https://api.openai.com/v1/completions"

var prompt = `\nConvert text to valid linux commands\n\ntext: `
var endPrompt = `\ncommand: `

var client = &http.Client{}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readRequestBody(req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}

func parseOpenAIResponse(resp *http.Response) (*OpenAIResponse, error) {
	ai := &OpenAIResponse{}
	err := json.NewDecoder(resp.Body).Decode(ai)
	return ai, err
}

func fzf(choices []string) {
	idx, err := fuzzyfinder.Find(
		choices,
		func(i int) string {
			return choices[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return choices[i]
		}))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", strings.Trim(choices[idx], "\n"))

}

func uniqueValuesFromList(list []Choice) []string {
	uniqueValues := make(map[string]bool)
	lst := []string{}
	for _, item := range list {
		uniqueValues[strings.Trim(strings.Trim(item.Text, "\n"), " ")] = true
	}

	for key := range uniqueValues {
		lst = append(lst, key)
	}

	return lst
}

func main() {
	app := &cli.App{
		Name:  "tellme",
		Usage: "text to command",
		Action: func(ctx *cli.Context) error {

			if len(ctx.Args().Slice()) == 0 {
				fmt.Println("Please provide a text to convert to a command")
				os.Exit(1)
			}
			text := strings.Join(ctx.Args().Slice(), " ")

			body := []byte(`{
				"model": "text-davinci-002",
				"prompt": "` + prompt + text + endPrompt + `",
				"frequency_penalty": 1,
				"temperature": 0.5,
				"top_p": 1,
				"n": 5,
				"max_tokens": 250
			}`)

			req, err := http.NewRequest("POST", openaiURI, bytes.NewBuffer(body))
			checkError(err)

			// readRequestBody(req)
			req.Header.Add("Authorization", "Bearer "+os.Getenv("TELLME_OPENAI_TOKEN"))
			req.Header.Set("Content-Type", "application/json; charset=UTF-8")

			res, err := client.Do(req)
			checkError(err)

			aires, err := parseOpenAIResponse(res)
			checkError(err)

			lst := uniqueValuesFromList(aires.Choices)
			fzf(lst)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
