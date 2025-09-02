// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	gobedrock "entest/gobedrock/bedrock"
	"fmt"
	"log"
	"net/http"
	"os"
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
	// Initialize X-Ray SDK
	fmt.Println("init and create an opensearch client")
	gobedrock.InitializeXRay()

	// create handler multiplexer
	mux := http.NewServeMux()

	// frontend claude haiku
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, error := os.ReadFile("./static/claude-haiku.html")
		if error != nil {
			fmt.Println(error)
		}
		w.Write(content)
	})

	// backend claude haiku
	mux.HandleFunc("/bedrock-haiku", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gobedrock.HandleBedrockClaude3HaikuChat(w, r, BedrockClient)
		}
	})

	// knowledge based retrieve frontend
	mux.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		content, error := os.ReadFile("./static/retrieve.html")
		if error != nil {
			fmt.Println(error)
		}
		w.Write(content)
	})

	// knowledge based retrieve backend
	mux.HandleFunc("/knowledge-base-retrieve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gobedrock.HandleRetrieve(w, r, BedrockAgentRuntimeClient)
		}
	})

	// knowledge based retrieve frontend
	mux.HandleFunc("/retrieve-generate", func(w http.ResponseWriter, r *http.Request) {
		content, error := os.ReadFile("./static/retrieve-and-generate.html")
		if error != nil {
			fmt.Println(error)
		}
		w.Write(content)

	})

	// knowledge based retrieve backend
	mux.HandleFunc("/knowledge-base-retrieve-and-generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gobedrock.HandleRetrieveAndGenerate(w, r, BedrockAgentRuntimeClient)
		}
	})

	// handle aoss index frontend
	mux.HandleFunc("/aoss-index", func(w http.ResponseWriter, r *http.Request) {
		content, error := os.ReadFile("./static/aoss-index.html")
		if error != nil {
			fmt.Println(error)
		}
		w.Write(content)

	})

	// handle index to aoss
	mux.HandleFunc("/aoss-index-backend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gobedrock.HandleAOSSIndex(w, r, AOSSClient, BedrockClient)
		}

	})

	// handle aoss query frontend
	mux.HandleFunc("/aoss-query", func(w http.ResponseWriter, r *http.Request) {
		content, error := os.ReadFile("./static/aoss-query.html")
		if error != nil {
			fmt.Println(error)
		}
		w.Write(content)

	})

	// handle query to aoss backend
	mux.HandleFunc("/aoss-query-backend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gobedrock.HandleAOSSQueryByTitle(w, r, AOSSClient, BedrockClient)
		}
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

	fmt.Println("Starting server on :3000 with AWS X-Ray tracing...")
	server.ListenAndServe()

}
