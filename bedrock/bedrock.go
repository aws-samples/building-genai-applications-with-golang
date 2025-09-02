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
	"time"

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



func HandleBedrockClaude3HaikuChat(w http.ResponseWriter, r *http.Request, BedrockClient *bedrockruntime.Client) {
	// Start monitoring: capture request start time for total latency calculation
	startTime := time.Now()
	// Generate unique request ID using nanosecond timestamp for log correlation
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	
	// Log request start in JSON format for CloudWatch Insights parsing
	fmt.Printf(`{"timestamp":"%s","request_id":"%s","event":"bedrock_request_start","model_id":"%s"}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, MODEL_ID, "\n")

	// list of messages sent from frontend client
	var request FrontEndRequest

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		// Log parsing errors with request context for debugging
		fmt.Printf(`{"timestamp":"%s","request_id":"%s","event":"request_parse_error","model_id":"%s","error":"%s"}%s`, 
			time.Now().UTC().Format(time.RFC3339), requestID, MODEL_ID, error.Error(), "\n")
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

	// Start timing Bedrock API call specifically (separate from total request time)
	bedrockStartTime := time.Now()
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
		// Log Bedrock API errors with latency - captures throttling, quota exceeded, etc.
		fmt.Printf(`{"timestamp":"%s","request_id":"%s","event":"bedrock_error","model_id":"%s","error":"%s","latency_ms":%d}%s`, 
			time.Now().UTC().Format(time.RFC3339), requestID, MODEL_ID, error.Error(), 
			time.Since(bedrockStartTime).Milliseconds(), "\n")
		fmt.Println(error)
		return
	}

	// Log successful Bedrock API response with time to first byte (TTFB)
	// This measures how long it took Bedrock to start streaming the response
	fmt.Printf(`{"timestamp":"%s","request_id":"%s","event":"bedrock_stream_start","model_id":"%s","latency_ms":%d}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, MODEL_ID,
		time.Since(bedrockStartTime).Milliseconds(), "\n")

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

	// Log request completion with total end-to-end latency
	// This includes request parsing + Bedrock API call + response streaming
	fmt.Printf(`{"timestamp":"%s","request_id":"%s","event":"request_complete","model_id":"%s","total_latency_ms":%d}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, MODEL_ID,
		time.Since(startTime).Milliseconds(), "\n")
}


