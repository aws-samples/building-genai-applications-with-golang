// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// generateRequestID creates a unique request ID for tracing
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func HandleRetrieve(w http.ResponseWriter, r *http.Request, client *bedrockagentruntime.Client) {

	// parse user messages
	type Content struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type Message struct {
		Role    string    `json:"role"`
		Content []Content `json:"content"`
	}

	var request struct {
		Messages []Message `json:"messages"`
	}

	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		fmt.Println(error)
	}

	messages := request.Messages

	// pop the last message as user question
	userQuestion := messages[len(messages)-1].Content[0].Text

	// invoke bedrock agent runtime to retreive opensearch
	output, error := client.Retrieve(
		context.TODO(),
		&bedrockagentruntime.RetrieveInput{
			KnowledgeBaseId: aws.String(KNOWLEDGE_BASE_ID),
			RetrievalQuery: &types.KnowledgeBaseQuery{
				Text: aws.String(userQuestion),
			},
			RetrievalConfiguration: &types.KnowledgeBaseRetrievalConfiguration{
				VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
					NumberOfResults: aws.Int32(KNOWLEDGE_BASE_NUMBER_OF_RESULT),
				},
			},
		},
	)

	if error != nil {
		fmt.Println(error)
	}

	if error != nil {
		fmt.Println(error)
	}

	json.NewEncoder(w).Encode(output)
}

func HandleRetrieveAndGenerate(w http.ResponseWriter, r *http.Request, client *bedrockagentruntime.Client) {
	// Start X-Ray segment for the entire request
	ctx := r.Context()
	var seg *xray.Segment
	if os.Getenv("USE_XRAY_SDK") == "true" {
		ctx, seg = xray.BeginSegment(ctx, "bedrock.retrieve_and_generate")
		defer seg.Close(nil)
	}
	
	// Start timing for latency monitoring
	startTime := time.Now()
	
	// Generate unique request ID for tracing
	requestID := generateRequestID()
	
	// Add metadata to X-Ray segment
	if seg != nil {
		seg.AddMetadata("request_id", requestID)
		seg.AddMetadata("knowledge_base_id", KNOWLEDGE_BASE_ID)
		seg.AddMetadata("model_arn", KNOWLEDGE_BASE_MODEL_ID)
	}
	
	// Get trace ID for logging
	traceID := requestID
	if seg != nil {
		traceID = seg.TraceID
	}
	
	// Log request start with structured logging for CloudWatch Insights
	fmt.Printf(`{"timestamp":"%s","level":"INFO","request_id":"%s","trace_id":"%s","event":"beYou can now use the CloudWatch agent to collect metrics, logs and traces from Amazon EC2 instances and on-premise servers. CloudWatch agent version 1.300025.0 and later can collect traces from OpenTelemetry or X-Ray client SDKs, and send them to X-Ray. Using the CloudWatch agent instead of the AWS Distro for OpenTelemetry (ADOT) Collector or X-Ray daemon to collect traces can help you reduce the number of agents that you manage. See the CloudWatch agent topic in the CloudWatch User Guide for more information.
	drock_retrieve_generate_start","knowledge_base_id":"%s"}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, traceID, KNOWLEDGE_BASE_ID, "\n")

	// parse user messages
	type Content struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type Message struct {
		Role    string    `json:"role"`
		Content []Content `json:"content"`
	}

	var request struct {
		Messages []Message `json:"messages"`
	}

	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		// Log parsing error with request ID and trace info
		if seg != nil {
			seg.AddError(error)
		}
		fmt.Printf(`{"timestamp":"%s","level":"ERROR","request_id":"%s","trace_id":"%s","event":"request_parse_error","error":"%s"}%s`, 
			time.Now().UTC().Format(time.RFC3339), requestID, traceID, error.Error(), "\n")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	messages := request.Messages

	// pop the last message as user question
	userQuestion := messages[len(messages)-1].Content[0].Text
	
	// Add question metadata to X-Ray segment
	if seg != nil {
		seg.AddMetadata("question_length", len(userQuestion))
		seg.AddMetadata("messages_count", len(messages))
	}
	
	// Log user question (truncated for privacy)
	questionPreview := userQuestion
	if len(questionPreview) > 100 {
		questionPreview = questionPreview[:100] + "..."
	}
	fmt.Printf(`{"timestamp":"%s","level":"INFO","request_id":"%s","trace_id":"%s","event":"user_question","question_length":%d,"question_preview":"%s"}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, traceID, len(userQuestion), questionPreview, "\n")

	// Create X-Ray subsegment for Bedrock API call
	var bedrockSeg *xray.Segment
	if seg != nil {
		ctx, bedrockSeg = xray.BeginSubsegment(ctx, "bedrock.retrieve_and_generate_api")
		defer bedrockSeg.Close(nil)
		
		bedrockSeg.AddMetadata("aws.service", "bedrock-agent-runtime")
		bedrockSeg.AddMetadata("aws.operation", "RetrieveAndGenerate")
		bedrockSeg.AddMetadata("knowledge_base_id", KNOWLEDGE_BASE_ID)
		bedrockSeg.AddMetadata("model_arn", KNOWLEDGE_BASE_MODEL_ID)
		bedrockSeg.AddMetadata("retrieval_number_of_results", int(KNOWLEDGE_BASE_NUMBER_OF_RESULT))
	}
	
	// Record time before Bedrock call
	bedrockStartTime := time.Now()

	// invoke bedrock agent runtime to retrieve and generate
	output, error := client.RetrieveAndGenerate(
		ctx,
		&bedrockagentruntime.RetrieveAndGenerateInput{
			Input: &types.RetrieveAndGenerateInput{
				Text: aws.String(userQuestion),
			},
			RetrieveAndGenerateConfiguration: &types.RetrieveAndGenerateConfiguration{
				Type: types.RetrieveAndGenerateTypeKnowledgeBase,
				KnowledgeBaseConfiguration: &types.KnowledgeBaseRetrieveAndGenerateConfiguration{
					KnowledgeBaseId: aws.String(KNOWLEDGE_BASE_ID),
					ModelArn:        aws.String(KNOWLEDGE_BASE_MODEL_ID),
					RetrievalConfiguration: &types.KnowledgeBaseRetrievalConfiguration{
						VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
							NumberOfResults: aws.Int32(KNOWLEDGE_BASE_NUMBER_OF_RESULT),
						},
					},
				},
			},
		},
	)

	// Calculate Bedrock API latency
	bedrockLatency := time.Since(bedrockStartTime).Milliseconds()
	if bedrockSeg != nil {
		bedrockSeg.AddMetadata("duration_ms", bedrockLatency)
	}

	if error != nil {
		// Record error in X-Ray segments
		if bedrockSeg != nil {
			bedrockSeg.AddError(error)
		}
		if seg != nil {
			seg.AddError(error)
		}
		
		// Log Bedrock API error with latency and trace info
		fmt.Printf(`{"timestamp":"%s","level":"ERROR","request_id":"%s","trace_id":"%s","event":"bedrock_api_error","error":"%s","bedrock_latency_ms":%d}%s`, 
			time.Now().UTC().Format(time.RFC3339), requestID, traceID, error.Error(), bedrockLatency, "\n")
		
		http.Error(w, "Bedrock API error", http.StatusInternalServerError)
		return
	}

	// Log successful Bedrock response with metrics
	responseLength := 0
	citationsCount := 0
	sessionID := ""
	
	if output.Output != nil && output.Output.Text != nil {
		responseLength = len(*output.Output.Text)
	}
	if output.Citations != nil {
		citationsCount = len(output.Citations)
	}
	if output.SessionId != nil {
		sessionID = *output.SessionId
	}

	// Add response metadata to X-Ray segments
	if bedrockSeg != nil {
		bedrockSeg.AddMetadata("response_length", responseLength)
		bedrockSeg.AddMetadata("citations_count", citationsCount)
		bedrockSeg.AddMetadata("session_id", sessionID)
	}
	if seg != nil {
		seg.AddMetadata("response_length", responseLength)
		seg.AddMetadata("citations_count", citationsCount)
		seg.AddMetadata("session_id", sessionID)
	}

	fmt.Printf(`{"timestamp":"%s","level":"INFO","request_id":"%s","trace_id":"%s","event":"bedrock_api_success","bedrock_latency_ms":%d,"response_length":%d,"citations_count":%d,"session_id":"%s"}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, traceID, bedrockLatency, responseLength, citationsCount, sessionID, "\n")

	// write output to client
	json.NewEncoder(w).Encode(output)

	// Calculate total request latency
	totalLatency := time.Since(startTime).Milliseconds()
	if seg != nil {
		seg.AddMetadata("total_duration_ms", totalLatency)
	}
	
	// Log request completion with full metrics and trace info
	fmt.Printf(`{"timestamp":"%s","level":"INFO","request_id":"%s","trace_id":"%s","event":"bedrock_retrieve_generate_complete","total_latency_ms":%d,"bedrock_latency_ms":%d,"knowledge_base_id":"%s","model_arn":"%s"}%s`, 
		time.Now().UTC().Format(time.RFC3339), requestID, traceID, totalLatency, bedrockLatency, KNOWLEDGE_BASE_ID, KNOWLEDGE_BASE_MODEL_ID, "\n")
}
