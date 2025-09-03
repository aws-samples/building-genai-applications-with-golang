// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Global Claude model parameters
const MAX_TOKENS_TO_SAMPLE = 2048
const ANTHROPIC_VERSION = "bedrock-2023-05-31"
const TEMPERATURE = 0.9

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

type Response struct {
	Completion string `json:"completion"`
}

type Query struct {
	Topic string `json:"topic"`
}

func HandleBedrockClaude3HaikuChat(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client) {

	// list of messages sent from frontend client
	var request FrontEndRequest

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		panic(error)
	}

	messages := request.Messages

	fmt.Println(messages)

	payload := RequestBodyClaude3{
		MaxTokensToSample: MAX_TOKENS_TO_SAMPLE,
		AnthropicVersion:  ANTHROPIC_VERSION,
		Temperature:       TEMPERATURE,
		Messages:          messages,
	}

	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		fmt.Println(error)
	}

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String(MODEL_ID),
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

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Println(error)
			}

			// stream to client
			// fmt.Println(resp.Delta.Text)
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


func HandleHaikuImageAnalyzer(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client) {

	// allow cros
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	type Message struct {
		Role    string        `json:"role"`
		Content []interface{} `json:"content"`
	}

	type Request struct {
		Messages []Message `json:"messages"`
	}

	type RequestBodyClaude3 struct {
		MaxTokensToSample int       `json:"max_tokens"`
		Temperature       float64   `json:"temperature,omitempty"`
		AnthropicVersion  string    `json:"anthropic_version"`
		Messages          []Message `json:"messages"`
	}

	var request Request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		panic(error)
	}

	messages := request.Messages

	payload := RequestBodyClaude3{
		MaxTokensToSample: MAX_TOKENS_TO_SAMPLE,
		AnthropicVersion:  ANTHROPIC_VERSION,
		Temperature:       TEMPERATURE,
		Messages:          messages,
	}

	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		fmt.Println(error)
	}

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String(MODEL_ID),
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

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(resp.Delta.Text)
			var tpl = template.Must(template.New("tpl").Parse(resp.Delta.Text))
			tpl.Execute(w, nil)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				fmt.Print("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)

		default:
			fmt.Println("union is nil or unknown type")
		}
	}
}
