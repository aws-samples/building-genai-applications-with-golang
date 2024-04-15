// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	gobedrock "entest/gobedrock/bedrock"
	"fmt"
	"log"
	"net/http"
	"html/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
	"github.com/rs/cors"
)

// opensearch severless client
var AOSSClient *opensearch.Client

// bedrock runtime client
var BedrockClient *bedrockruntime.Client

// bedrock agent runtime client
var BedrockAgentRuntimeClient *bedrockagentruntime.Client

// create an init function to initializing opensearch client
func init() {

	//
	fmt.Println("init and create an opensearch client")

	// load aws credentials from profile demo using config
	awsCfg1, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(gobedrock.BEDROCK_REGION),
	)

	if err != nil {
		log.Fatal(err)
	}

	awsCfg2, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(gobedrock.AOSS_REGION),
	)

	if err != nil {
		log.Fatal(err)
	}

	// create a aws request signer using requestsigner
	signer, err := requestsigner.NewSignerWithService(awsCfg2, "aoss")

	if err != nil {
		log.Fatal(err)
	}

	// create an opensearch client using opensearch package
	AOSSClient, err = opensearch.NewClient(opensearch.Config{
		Addresses: []string{gobedrock.AOSS_ENDPOINT},
		Signer:    signer,
	})

	if err != nil {
		log.Fatal(err)
	}

	// create bedrock runtime client
	BedrockClient = bedrockruntime.NewFromConfig(awsCfg1)

	// create bedrock agent runtime client
	BedrockAgentRuntimeClient = bedrockagentruntime.NewFromConfig(awsCfg1)

}

func main() {

	// create handler multiplexer
	mux := http.NewServeMux()

	// frontend claude haiku
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/claude-haiku.html")
	})

	// backend claude haiku
	mux.HandleFunc("/bedrock-haiku", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleBedrockClaude3HaikuChat(w, r, BedrockClient)
	})

	// bedrock frontend for image analyzer
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/image.html")
	})

	// bedrock backend to analyze image
	mux.HandleFunc("/claude-haiku-image", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleHaikuImageAnalyzer(w, r, BedrockClient)
	})

	// frontend claude2
	mux.HandleFunc("/claude2", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/claude2.html")
	})

	// backend claude2
	mux.HandleFunc("/bedrock-stream", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleBedrockClaude2Chat(w, r, BedrockClient)
	})

	// handle aoss frontend
	mux.HandleFunc("/aoss", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			name := r.FormValue("name")
			var tpl = template.Must(template.New("tpl").Parse(name))
			tpl.Execute(w, nil)
			// another way and client parse it
			// json.NewEncoder(w).Encode(resp)
		}

		if r.Method == "GET" {
			http.ServeFile(w, r, "./static/opensearch.html")
		}
	})

	// handle query to aoss
	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleAOSSQuery(w, r, AOSSClient, BedrockClient)
	})

	// frontend for rag
	mux.HandleFunc("/rag", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/rag.html")
	})

	// backend for rag
	mux.HandleFunc("/rag-query", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleRagQueryClaude3(w, r, BedrockClient, AOSSClient)
	})

	// magic mirror frontend
	mux.HandleFunc("/mirror", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/mirror.html")
	})

	// knowledge based retrieve frontend
	mux.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/retrieve.html")
	})

	// knowledge based retrieve backend
	mux.HandleFunc("/knowledge-base-retrieve", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleRetrieve(w, r, BedrockAgentRuntimeClient)
	})

	// knowledge based retrieve frontend
	mux.HandleFunc("/retrieve-generate", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/retrieve-and-generate.html")
	})

	// knowledge based retrieve backend
	mux.HandleFunc("/knowledge-base-retrieve-and-generate", func(w http.ResponseWriter, r *http.Request) {
		gobedrock.HandleRetrieveAndGenerate(w, r, BedrockAgentRuntimeClient)
	})

	// allow cors
	handler := cors.AllowAll().Handler(mux)

	// create a http server using http
	server := http.Server{
		Addr:           ":3000",
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server.ListenAndServe()

}
