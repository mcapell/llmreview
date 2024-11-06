package main

import (
	"fmt"

	"github.com/mcapell/llmreview/question"
)

func main() {
	questions, err := question.ParseQuestionsFromPath("./data/")
	if err != nil {
		panic(err)
	}

	for _, q := range questions {
		fmt.Printf("parsed question: %+v\n", q)
	}
}
