// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type IndexItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Text  string `json:"text"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

type Hits struct {
	Hits []map[string]interface{} `json:"hits"`
}

type AossResponse struct {
	Hits Hits `json:"hits"`
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

func QueryAOSSByVector(vec []float64, AOSSClient *opensearch.Client) (*opensearchapi.Response, error) {

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

	search := opensearchapi.SearchRequest{
		Index: []string{AOSS_NOTE_APP_INDEX_NAME},
		Body:  content,
	}

	response, error := search.Do(context.Background(), AOSSClient)

	if error != nil {
		log.Fatal(error)
	}

	return response, nil

}

func HandleAOSSQueryByVector(w http.ResponseWriter, r *http.Request, AOSSClient *opensearch.Client, BedrockClient *bedrockruntime.Client) {

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

	// convert query to embedding vector
	vec, error := GetEmbedVector(query, BedrockClient)

	if error != nil {
		fmt.Println(error)
	}

	// query opensearch
	response, error := QueryAOSSByVector(vec, AOSSClient)

	if error != nil {
		fmt.Println(error)
	}

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// write answer to response
	json.NewEncoder(w).Encode(map[string]interface{}{"Result": string(respBytes)})
}

func QueryOpenSearchByTitle(AOSSClient *opensearch.Client, title string) (*opensearchapi.Response, error) {

	content := strings.NewReader(fmt.Sprintf(`{
		"size": 10,
		"query": {
			"multi_match": {
				"query": "%s",
				"fields": ["title"]
			}
	}
}`, title))

	search := opensearchapi.SearchRequest{
		Index: []string{AOSS_NOTE_APP_INDEX_NAME},
		Body:  content,
	}

	response, error := search.Do(context.Background(), AOSSClient)

	if error != nil {
		fmt.Println(error)
	}

	return response, nil

}

func HandleAOSSQueryByTitle(w http.ResponseWriter, r *http.Request, AOSSClient *opensearch.Client, BedrockClient *bedrockruntime.Client) {

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

	// query opensearh match by title
	response, error := QueryOpenSearchByTitle(AOSSClient, query)

	if error != nil {
		fmt.Println(error)
	}

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// write answer to response
	json.NewEncoder(w).Encode(map[string]interface{}{"Result": string(respBytes)})
}

func IndexVectorOpenSearch(AOSSClient *opensearch.Client, BedrockClient *bedrockruntime.Client, item IndexItem) (*opensearchapi.Response, error) {

	// get embedding vector
	vec, error := GetEmbedVector(item.Text, BedrockClient)

	if error != nil {
		fmt.Println(error)
	}

	// convert vector of number to string
	vecStr := make([]string, len(vec))

	// convert array float to string
	for k, v := range vec {

		if k < len(vec)-1 {
			vecStr[k] = fmt.Sprint(v) + ","
		} else {
			vecStr[k] = fmt.Sprint(v)
		}
	}

	// body request for indexing opensearch
	body := strings.NewReader(fmt.Sprintf(`{
		"title": "%s",
		"link": "%s",
		"text": "%s",
		"vector_field": %s
	}`, item.Title, item.Link, item.Text, vecStr))

	index := opensearchapi.IndexRequest{
		Index: AOSS_NOTE_APP_INDEX_NAME,
		Body:  body,
	}

	// index into opensearch
	response, error := index.Do(context.Background(), AOSSClient)

	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(response)

	return response, nil

}

func HandleAOSSIndex(w http.ResponseWriter, r *http.Request, AOSSClient *opensearch.Client, BedrockClient *bedrockruntime.Client) {

	// data struct of request
	var request struct {
		Title string `json:"title"`
		Text  string `json:"text"`
		Link  string `json:"link"`
	}

	// parse request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		fmt.Println(error)
	}

	// index into opensearch
	response, error := IndexVectorOpenSearch(AOSSClient, BedrockClient, IndexItem{Title: request.Title, Link: request.Link, Text: request.Text})

	if error != nil {
		fmt.Println(error)
	}

	//
	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// write json encoding to response
	json.NewEncoder(w).Encode(map[string]interface{}{"Result": string(respBytes)})

}
