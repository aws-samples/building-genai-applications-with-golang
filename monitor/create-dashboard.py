# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import boto3
import json

# parameter
DASHBOARD_NAME = "AmazonClaude35Sonnet"
REGION = "us-west-2"
MODEL_ID = "us.anthropic.claude-3-5-sonnet-20240620-v1:0"

# cloudwatch client
client = boto3.client("cloudwatch", region_name=REGION)
# dashboard boy
dashboard_body = {
    "widgets": [
        {
            "height": 7,
            "width": 7,
            "y": 0,
            "x": 0,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "OutputTokenCount",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ],
                    [".", "InputTokenCount", ".", ".", {"region": REGION}],
                ],
                "view": "timeSeries",
                "stacked": False,
                "region": REGION,
                "period": 60,
                "stat": "Sum",
                "title": "TokenCountPerMinute",
            },
        },
        {
            "height": 7,
            "width": 7,
            "y": 0,
            "x": 7,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "Invocations",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ]
                ],
                "view": "timeSeries",
                "stacked": False,
                "region": REGION,
                "period": 60,
                "stat": "Sum",
                "title": "InvocationPerMinute",
            },
        },
        {
            "height": 7,
            "width": 7,
            "y": 0,
            "x": 14,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "Invocations",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ]
                ],
                "view": "gauge",
                "yAxis": {"left": {"min": 0, "max": 100}},
                "region": REGION,
                "period": 60,
                "stat": "Sum",
                "title": "InvocationPerMinute",
            },
        },
        {
            "height": 7,
            "width": 7,
            "y": 7,
            "x": 7,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "InvocationThrottles",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ]
                ],
                "view": "timeSeries",
                "stacked": False,
                "region": REGION,
                "period": 60,
                "stat": "Sum",
                "title": "InvocationThrottlesPerMinute",
            },
        },
        {
            "height": 7,
            "width": 7,
            "y": 7,
            "x": 14,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "InvocationClientErrors",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ]
                ],
                "view": "timeSeries",
                "stacked": False,
                "region": REGION,
                "period": 60,
                "stat": "Sum",
            },
        },
        {
            "height": 7,
            "width": 7,
            "y": 7,
            "x": 0,
            "type": "metric",
            "properties": {
                "metrics": [
                    [
                        "AWS/Bedrock",
                        "InvocationLatency",
                        "ModelId",
                        MODEL_ID,
                        {"region": REGION},
                    ]
                ],
                "view": "timeSeries",
                "stacked": False,
                "region": REGION,
                "period": 60,
                "stat": "Average",
                "title": "AvgInvocationLatency",
            },
        },
    ]
}
#
print(json.dumps(dashboard_body))
#
response = client.put_dashboard(
    DashboardName=DASHBOARD_NAME, DashboardBody=json.dumps(dashboard_body)
)
