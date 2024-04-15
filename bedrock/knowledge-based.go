// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

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

	fmt.Println(messages)

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

	// for k, v := range output.RetrievalResults {
	// 	fmt.Println(k)
	// 	fmt.Println("=======================================")
	// 	fmt.Println(*v.Content.Text)
	// }

	// parse output to []byte and return client
	// outputJson, error := json.Marshal(output)

	if error != nil {
		fmt.Println(error)
	}

	json.NewEncoder(w).Encode(output)

}

func HandleRetrieveAndGenerate(w http.ResponseWriter, r *http.Request, client *bedrockagentruntime.Client) {

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

	fmt.Println(messages)

	// pop the last message as user question
	userQuestion := messages[len(messages)-1].Content[0].Text

	// invoke bedrock agent runtime to retrieve and generate
	output, error := client.RetrieveAndGenerate(
		context.TODO(),
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

	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(*output.Output.Text)

	// parse output to []byte
	// outputJson, error := json.Marshal(output)

	if error != nil {
		fmt.Println(error)
	}

	// write output to client
	json.NewEncoder(w).Encode(output)

}
