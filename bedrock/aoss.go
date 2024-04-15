// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

type Hits struct {
	Hits []map[string]interface{} `json:"hits"`
}

type AossResponse struct {
	Hits Hits `json:"hits"`
}

func QueryAOSS(vec []float64, AOSSClient *opensearch.Client) ([]string, error) {

	// let query get all item in an index

	// content := strings.NewReader(`{
	//     "size": 10,
	//     "query": {
	//         "match_all": {}
	//         }
	// }`)

	vecStr := make([]string, len(vec))

	// convert array float to string
	for k, v := range vec {

		if k < len(vec)-1 {
			vecStr[k] = fmt.Sprint(v) + ","
		} else {
			vecStr[k] = fmt.Sprint(v)
		}

	}

	// create request body to titan model
	content := strings.NewReader(fmt.Sprintf(`{
		"size": 5,
		"query": {
			"knn": {
				"vector_field": {
					"vector": %s,
					"k": 5
				}
			}
		}
	}`, vecStr))

	// fmt.Println(content)

	search := opensearchapi.SearchRequest{
		Index: []string{"demo"},
		Body:  content,
	}

	searchResponse, err := search.Do(context.Background(), AOSSClient)

	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(searchResponse)

	var answer AossResponse

	json.NewDecoder(searchResponse.Body).Decode(&answer)

	// first := answer.Hits.Hits[0]

	// fmt.Printf("id: %s\n, index: %s\n, text: %s", first["_id"], first["_index"], first["_source"].(map[string]interface{})["text"])

	// fmt.Println(answer.Hits.Hits[0]["_id"])

	queryResult := answer.Hits.Hits[0]["_source"].(map[string]interface{})["text"]

	if queryResult == nil {
		return []string{"nil"}, nil
	}

	// extract hint text only
	hits := []string{}

	for k, v := range answer.Hits.Hits {

		if k >= 0 {
			hits = append(hits, v["_source"].(map[string]interface{})["text"].(string))
		}

	}

	return hits, nil

	// return fmt.Sprint(queryResult), nil

}

func GetEmbedVector(question string, BedrockClient *bedrockruntime.Client) ([]float64, error) {

	// create request body to titan model
	body := map[string]interface{}{
		"inputText": question,
	}
	bodyJson, err := json.Marshal(body)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// invoke bedrock titan model to convert string to embedding vector
	response, error := BedrockClient.InvokeModel(
		context.Background(),
		&bedrockruntime.InvokeModelInput{
			Body:        []byte(bodyJson),
			ModelId:     aws.String("amazon.titan-embed-text-v1"),
			ContentType: aws.String("application/json"),
		},
	)

	if error != nil {
		fmt.Println(error)
		return nil, error
	}

	// assert response to map
	var embedResponse map[string]interface{}

	error = json.Unmarshal(response.Body, &embedResponse)

	if error != nil {
		fmt.Println(error)
		return nil, error
	}

	// assert response to array
	slice, ok := embedResponse["embedding"].([]interface{})

	if !ok {
		fmt.Println(ok)
	}

	// assert to array of float64
	values := make([]float64, len(slice))

	for k, v := range slice {
		values[k] = float64(v.(float64))
	}

	return values, nil
}

func HandleAOSSQuery(w http.ResponseWriter, r *http.Request, AOSSClient *opensearch.Client, BedrockClient *bedrockruntime.Client) {

	// data struct of request
	var request struct {
		Query string `json:"query"`
	}

	// parse user query from request
	var query string
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		fmt.Println(error)
	}

	query = request.Query
	fmt.Println(query)

	// convert query to embedding vector
	vec, error := GetEmbedVector(query, BedrockClient)

	if error != nil {
		fmt.Println(error)
	}

	// query opensearch
	answers, error := QueryAOSS(vec, AOSSClient)

	if error != nil {
		fmt.Println(error)
	}

	// write answer to response
	json.NewEncoder(w).Encode(struct {
		Messages []string `json:"Messages"`
	}{Messages: answers})
}


