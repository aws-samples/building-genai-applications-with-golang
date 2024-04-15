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
)

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

// claude2 data type
type Request struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type Response struct {
	Completion string `json:"completion"`
}

type HelloHandler struct{}

type Query struct {
	Topic string `json:"topic"`
}

func HandleBedrockClaude2Chat(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client) {

	const claudePromptFormat = "\n\nHuman: %s\n\nAssistant:"

	var query Query
	var message string

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&query)

	if error != nil {
		message = "how to learn japanese as quick as possible?"
		panic(error)
	}

	message = query.Topic

	fmt.Println(message)

	prompt := "" + fmt.Sprintf(claudePromptFormat, message)

	payload := Request{
		Prompt:            prompt,
		MaxTokensToSample: 2048,
	}

	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		fmt.Println(error)
	}

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-v2"),
			ContentType: aws.String("application/json"),
		},
	)

	if error != nil {
		fmt.Println(error)
	}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//fmt.Println("payload", string(v.Value.Bytes))
			var resp Response
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Println(err)
			}

			// stream to client
			fmt.Println(resp.Completion)
			var tpl = template.Must(template.New("tpl").Parse(resp.Completion))
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
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.9,
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
				fmt.Println(error)
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

func HandleHaikuImageAnalyzer(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client) {

	// allow cros
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	// data type request
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

	// data type response
	// type ResponseContent struct {
	// 	Type string `json:"type"`
	// 	Text string `json:"text"`
	// }

	// type Response struct {
	// 	Content []ResponseContent `json:"content"`
	// }

	// parse request
	var request Request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		panic(error)
	}

	// fmt.Println(request)

	// payload for bedrock claude3 haikue
	messages := request.Messages

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.9,
		Messages:          messages,
	}

	// convert payload struct to bytes
	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		fmt.Println(error)
	}

	// fmt.Println("invoke bedrock ...")

	// invoke bedrock claude3 haiku
	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		},
	)

	// response
	// var response Response
	// json.NewDecoder(bytes.NewReader(output.Body)).Decode(&response)
	// fmt.Println(response)

	if error != nil {
		fmt.Println(error)
	}

	// stream result to client
	for event := range output.GetStream().Events() {

		// fmt.Println(event)

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			// fmt.Println("payload", string(v.Value.Bytes))

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
				fmt.Print("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)

		default:
			fmt.Println("union is nil or unknown type")
		}
	}
}
