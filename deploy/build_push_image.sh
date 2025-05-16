#!/bin/bash
set -e

# Default values
REGION="us-west-2"
REPOSITORY_NAME="go-bedrock-app"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --region)
      REGION="$2"
      shift 2
      ;;
    --repository-name)
      REPOSITORY_NAME="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $key"
      echo "Usage: $0 [--repository-name REPOSITORY_NAME] [--region REGION]"
      exit 1
      ;;
  esac
done

# Get AWS account ID
ACCOUNT=$(aws sts get-caller-identity | jq -r '.Account')

# Navigate to the project root directory
cd "$(dirname "$0")/.."

echo "Using AWS region: $REGION"
echo "Using ECR repository name: $REPOSITORY_NAME"
echo "Building Docker image: $REPOSITORY_NAME"

# Build the Docker image
sudo docker build -t $REPOSITORY_NAME .

echo "Logging in to AWS ECR in region $REGION"
aws ecr get-login-password --region $REGION | sudo docker login --username AWS --password-stdin $ACCOUNT.dkr.ecr.$REGION.amazonaws.com

# Get docker image ID
IMAGE_ID=$(sudo docker images -q $REPOSITORY_NAME:latest)

echo "Tagging image with ECR repository"
sudo docker tag $IMAGE_ID $ACCOUNT.dkr.ecr.$REGION.amazonaws.com/$REPOSITORY_NAME:latest

echo "Pushing image to ECR"
sudo docker push $ACCOUNT.dkr.ecr.$REGION.amazonaws.com/$REPOSITORY_NAME:latest

echo "Successfully built and pushed $REPOSITORY_NAME to ECR"
