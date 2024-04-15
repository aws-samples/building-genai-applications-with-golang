// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"html/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

func GetContext(currentQuestion string, BedrockClient *bedrockruntime.Client, AOSSClient *opensearch.Client) (string, error) {

	// convert question to vector
	vec, error := GetEmbedVector(currentQuestion, BedrockClient)

	if error != nil {
		return "no information in database", error
	}

	// query opensearch
	docs, error := QueryAOSS(vec, AOSSClient)

	if error != nil {
		return "no information in database", error
	}

	// concatenate all docs to a string
	var docsString string

	for i, doc := range docs {
		docsString += doc
		// context from the top 2 documents
		if i > 2 {
			break
		}
	}

	return docsString, nil
}

func HandleRagQueryClaude3(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client, AOSSClient *opensearch.Client) {

	// claude3 request data type
	type Content struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type Message struct {
		Role    string    `json:"role"`
		Content []Content `json:"content"`
	}

	type RequestBodyClaude3 struct {
		MaxTokensToSample int       `json:"max_tokens"`
		Temperature       float64   `json:"temperature,omitempty"`
		AnthropicVersion  string    `json:"anthropic_version"`
		Messages          []Message `json:"messages"`
	}

	// frontend request data type
	type FrontEndRequest struct {
		Messages []Message `json:"messages"`
	}

	// claude3 response data type
	type Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type ResponseClaude3 struct {
		Type  string `json:"type"`
		Index int    `json:"index"`
		Delta Delta  `json:"delta"`
	}

	// list of messages sent from frontend client
	var request FrontEndRequest

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		panic(error)
	}

	messages := request.Messages

	// fmt.Println(messages)

	// based on current user question let query opensearch to get context
	currentQuestion := messages[len(messages)-1].Content[0].Text

	docs, error := GetContext(currentQuestion, BedrockClient, AOSSClient)

	if error != nil {
		fmt.Println(error)
		docs = "no information in database"
	}

	// do some prompt engineering
	promptContext := fmt.Sprintf(" Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer. don't include harmful content <context> %s </context>. %s", docs, currentQuestion)

	// pass the context to messages of new prompt for claude
	// messages = append(messages, Message{
	// 	Role:    "user",
	// 	Content: []Content{{Type: "text", Text: promptContext}},
	// })

	// create a new messages
	newMessages := []Message{
		{
			Role:    "user",
			Content: []Content{{Type: "text", Text: promptContext}},
		},
	}

	fmt.Println(newMessages)

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.9,
		Messages:          newMessages,
	}

	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		fmt.Println(error)
	}

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		},
	)

	if error != nil {
		fmt.Println(error)
	}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//fmt.Println("payload", string(v.Value.Bytes))

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Println(err)
			}

			// stream to client
			fmt.Println(resp.Delta.Text)
			var tpl = template.Must(template.New("tpl").Parse(resp.Delta.Text))
			tpl.Execute(w, nil)
			// another way and client parse it
			// json.NewEncoder(w).Encode(resp)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				fmt.Println("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)

		default:
			fmt.Println("union is nil or unknown type")
		}
	}
}
