# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
import json
import boto3
from concurrent.futures import ThreadPoolExecutor
import time

# Number of concurrent user
NUM_THREAD = 5
# Sleep between invocation in a user
SLEEP_TIME = 1

# MODEL PARAMETER
# MODEL_ID = "us.anthropic.claude-3-haiku-20240307-v1:0"
MODEL_ID = "us.anthropic.claude-3-5-sonnet-20240620-v1:0"
# MODEL_ID = "us.anthropic.claude-3-haiku-20240307-v1:0"
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
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "text": PROMPT
                    }
                ]
            }
        ],
        "inferenceConfig": {
            "maxTokens": MAX_TOKEN,
            "temperature": TEMPERATURE,
            "topP": 0.9
        }
    }

    # Call the converse API
    try:
        response = client.converse(
            modelId=MODEL_ID,
            messages=request["messages"],
            inferenceConfig=request["inferenceConfig"]
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
    simple load test
    """
    count = 1
    while True:
        print(f"Thread: {id} request: {count}")
        try:
            converse_with_model()
        except Exception as e:
            print(f"ERROR: {e}")
        print(f"Thread: {id} completed request: {count}")
        count += 1
        time.sleep(SLEEP_TIME)


def simple_load_test_multi_thread():
    """
    simple load test multi thread
    """
    with ThreadPoolExecutor(max_workers=NUM_THREAD) as executor:
        for k in range(1, NUM_THREAD + 1):
            executor.submit(simple_load_test_single_thread, k)


if __name__ == "__main__":
    # converse_with_model()
    # simple_load_test_single_thread(id=1)
    simple_load_test_multi_thread()
