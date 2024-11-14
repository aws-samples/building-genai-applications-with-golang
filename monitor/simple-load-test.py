// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
import json
import boto3
from concurrent.futures import ThreadPoolExecutor
import time

# Number of concurrent user
NUM_THREAD = 5
# Sleep between invockation in a user
SLEEP_TIME = 1

# MODEL PARAMETER
# MODEL_ID = "us.anthropic.claude-3-haiku-20240307-v1:0"
MODEL_ID = "us.anthropic.claude-3-5-sonnet-20240620-v1:0"
MAX_TOKEN = 2048
TEMPERATURE = 0.9
ANTHROPIC_VERSION = "bedrock-2023-05-31"
RPOMPT = "tell me about amazon"

# Bedrock runtime client
client = boto3.client("bedrock-runtime", region_name="us-west-2")


def invoke_model():
    """
    invoke amazon claude
    """
    # anthropic claude message
    messages = [{"role": "user", "content": [{"type": "text", "text": RPOMPT}]}]
    # body
    body = json.dumps(
        {
            "anthropic_version": ANTHROPIC_VERSION,
            "max_tokens": MAX_TOKEN,
            "temperature": TEMPERATURE,
            "messages": messages,
        }
    )
    # invoke model
    response = client.invoke_model(
        body=body,
        modelId=MODEL_ID,
        accept="application/json",
        contentType="application/json",
    )
    # response
    response_body = json.loads(response.get("body").read())
    print(response_body)


def simple_load_test_single_thread(id):
    """
    simple load test
    """
    count = 1
    while True:
        print(f"Thread: {id} request: {count}")
        count += 1
        try:
            invoke_model()
        except:
            print("ERROR")
        print(f"Thread: {id} request: {count}")
        time.sleep(SLEEP_TIME)


def simple_load_test_multi_thread():
    """
    smple load test multi thread
    """
    with ThreadPoolExecutor(max_workers=NUM_THREAD) as executor:
        for k in range(1, NUM_THREAD + 1):
            executor.submit(simple_load_test_single_thread, k)


if __name__ == "__main__":
    # invoke_model()
    # simple_load_test_single_thread(id=1)
    simple_load_test_multi_thread()
