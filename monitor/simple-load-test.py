# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

"""
Claude Models - Request Rate Limits (Requests Per Minute)
=========================================================
| Model Family       | Model Version | Model ID                                           | Access Type       | Applied (RPM) | Default (RPM) |
|--------------------|---------------|----------------------------------------------------|-------------------|---------------|---------------|
| Claude 3.5 Sonnet  | 20240620-v1   | anthropic.claude-3-5-sonnet-20240620-v1:0          | ON_DEMAND         | 5             | 250           |
| Claude 3.5 Sonnet  | 20240620-v1   | us.anthropic.claude-3-5-sonnet-20240620-v1:0       | CROSS_REGION      | 10            | 500           |
| Claude 3.5 Sonnet  | 20241022-v2   | anthropic.claude-3-5-sonnet-20241022-v2:0          | ON_DEMAND         | 5             | 250           |
| Claude 3.5 Sonnet  | 20241022-v2   | us.anthropic.claude-3-5-sonnet-20241022-v2:0       | CROSS_REGION      | 10            | 500           |
| Claude Sonnet 4    | 20250514-v1   | us.anthropic.claude-sonnet-4-20250514-v1:0         | INFERENCE_PROFILE | 4             | 200           |
| Claude 3.5 Haiku   | 20241022-v1   | anthropic.claude-3-5-haiku-20241022-v1:0           | ON_DEMAND         | 20            | 1,000         |
| Claude 3.5 Haiku   | 20241022-v1   | anthropic.claude-3-5-haiku-20241022-v1:0           | LATENCY_OPTIMIZED | 100           | 100           |
| Claude 3.5 Haiku   | 20241022-v1   | us.anthropic.claude-3-5-haiku-20241022-v1:0        | CROSS_REGION      | 40            | 2,000         |
| Claude 3 Haiku     | 20240307-v1   | anthropic.claude-3-haiku-20240307-v1:0             | ON_DEMAND         | 20            | 1,000         |
| Claude 3 Haiku     | 20240307-v1   | us.anthropic.claude-3-haiku-20240307-v1:0          | CROSS_REGION      | 40            | 2,000         |

Claude Models - Token Rate Limits (Tokens Per Minute)
=====================================================
| Model Family       | Model Version | Access Type       | Input Tokens Applied | Input Tokens Default | Output Tokens Applied | Output Tokens Default |
|--------------------|---------------|-------------------|---------------------|--------------------|--------------------|---------------------|
| Claude 3.5 Sonnet  | 20240620-v1   | ON_DEMAND         | 2,000,000           | 2,000,000          | 2,000,000          | 2,000,000           |
| Claude 3.5 Sonnet  | 20240620-v1   | CROSS_REGION      | 4,000,000           | 4,000,000          | 4,000,000          | 4,000,000           |
| Claude 3.5 Sonnet  | 20241022-v2   | ON_DEMAND         | 2,000,000           | 2,000,000          | 2,000,000          | 2,000,000           |
| Claude 3.5 Sonnet  | 20241022-v2   | CROSS_REGION      | 4,000,000           | 4,000,000          | 4,000,000          | 4,000,000           |
| Claude Sonnet 4    | 20250514-v1   | CROSS_REGION      | 400,000             | 20,000,000         | 16,000             | 800,000             |
| Claude 3.5 Haiku   | 20241022-v1   | ON_DEMAND         | 2,000,000           | 2,000,000          | 2,000,000          | 2,000,000           |
| Claude 3.5 Haiku   | 20241022-v1   | LATENCY_OPTIMIZED | 500,000             | 500,000            | 500,000            | 500,000             |
| Claude 3.5 Haiku   | 20241022-v1   | CROSS_REGION      | 4,000,000           | 4,000,000          | 4,000,000          | 4,000,000           |
| Claude 3 Haiku     | 20240307-v1   | ON_DEMAND         | 4,000,000           | 200,000,000        | 160,000            | 8,000,000           |
| Claude 3 Haiku     | 20240307-v1   | CROSS_REGION      | 4,000,000           | 200,000,000        | 160,000            | 8,000,000           |
"""

import json
import boto3
from concurrent.futures import ThreadPoolExecutor
import time

# Number of concurrent user
NUM_THREAD = 5
# Sleep between invocation in a user
SLEEP_TIME = 1
# Maximum requests per thread (each user stops after this many requests)
MAX_REQUESTS_PER_THREAD = 10

# MODEL SELECTION - Change MODEL_ID to test different Claude models
# Refer to the tables above for Applied vs Default quota limits
# Default selection: Claude Sonnet 4 (most restrictive at 4 RPM - good for testing throttling)
MODEL_ID = "us.anthropic.claude-3-5-haiku-20241022-v1:0"  # Claude 3.5 Haiku - Applied: 100 RPM (latency-optimized)

MAX_TOKEN = 2048
TEMPERATURE = 0.9
PROMPT = "tell me about amazon"

# Bedrock runtime client
client = boto3.client("bedrock-runtime", region_name="us-west-2")


def converse_with_model():
    """
    Use the converse API to interact with Amazon Bedrock model
    """
    # Create the request body for converse API
    request = {
        "messages": [{"role": "user", "content": [{"text": PROMPT}]}],
        "inferenceConfig": {
            "maxTokens": MAX_TOKEN,
            "temperature": TEMPERATURE,
            "topP": 0.9,
        },
    }

    # Call the converse API
    try:
        response = client.converse(
            modelId=MODEL_ID,
            messages=request["messages"],
            inferenceConfig=request["inferenceConfig"],
        )

        # Process and print the response
        print(f"Response: {response['output']}")

        # You can access more details from the response if needed
        # print(f"Usage: {response['usage']}")

        return response
    except Exception as e:
        print(f"Error in converse API call: {e}")
        return None


def simple_load_test_single_thread(id):
    """
    simple load test - each thread stops after MAX_REQUESTS_PER_THREAD requests
    """
    count = 1
    while count <= MAX_REQUESTS_PER_THREAD:
        print(f"Thread: {id} request: {count}")
        try:
            converse_with_model()
        except Exception as e:
            print(f"ERROR: {e}")
        print(f"Thread: {id} completed request: {count}")
        count += 1
        if (
            count <= MAX_REQUESTS_PER_THREAD
        ):  # Only sleep if there are more requests to make
            time.sleep(SLEEP_TIME)

    print(f"Thread: {id} finished - completed {MAX_REQUESTS_PER_THREAD} requests")


def simple_load_test_multi_thread():
    """
    simple load test multi thread
    """
    with ThreadPoolExecutor(max_workers=NUM_THREAD) as executor:
        for k in range(1, NUM_THREAD + 1):
            executor.submit(simple_load_test_single_thread, k)


if __name__ == "__main__":
    print(f"Starting load test with model: {MODEL_ID}")
    simple_load_test_multi_thread()

