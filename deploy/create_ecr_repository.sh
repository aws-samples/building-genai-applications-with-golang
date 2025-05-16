#!/bin/bash

# Script to create an ECR repository
# Usage: ./create_ecr_repository.sh [--repository-name name] [--region region]
# Default repository name: go-bedrock-app
# Default region: us-west-2

# Set default values
DEFAULT_REPOSITORY_NAME="go-bedrock-app"
DEFAULT_REGION="us-west-2"

# Initialize with defaults
REPOSITORY_NAME=$DEFAULT_REPOSITORY_NAME
REGION=$DEFAULT_REGION

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --repository-name)
      REPOSITORY_NAME="$2"
      shift 2
      ;;
    --region)
      REGION="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--repository-name name] [--region region]"
      exit 1
      ;;
  esac
done

echo "Creating ECR repository '$REPOSITORY_NAME' in region '$REGION'..."

# Create the ECR repository
aws ecr create-repository \
    --repository-name $REPOSITORY_NAME \
    --region $REGION \
    --image-scanning-configuration scanOnPush=true

if [ $? -eq 0 ]; then
    echo "Repository '$REPOSITORY_NAME' created successfully in region '$REGION'."
    
    # Get the repository URI
    REPOSITORY_URI=$(aws ecr describe-repositories \
        --repository-names $REPOSITORY_NAME \
        --region $REGION \
        --query 'repositories[0].repositoryUri' \
        --output text)
    
    echo "Repository URI: $REPOSITORY_URI"
    echo ""
    echo "To push images to this repository:"
    echo "1. Authenticate Docker to your ECR registry:"
    echo "   aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $REPOSITORY_URI"
    echo ""
    echo "2. Build your Docker image:"
    echo "   docker build -t $REPOSITORY_NAME ."
    echo ""
    echo "3. Tag your image:"
    echo "   docker tag $REPOSITORY_NAME:latest $REPOSITORY_URI:latest"
    echo ""
    echo "4. Push the image to ECR:"
    echo "   docker push $REPOSITORY_URI:latest"
else
    echo "Failed to create repository. Check if it already exists or if you have sufficient permissions."
    exit 1
fi
