#!/bin/bash

# Set your AWS account ID and region
export AWS_ACCOUNT_ID=380206744475
export AWS_REGION=ap-southeast-2
export IMAGE_NAME=wcp-detrack-monthly-report

# Authenticate Docker to ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Create ECR repository
aws ecr create-repository \
  --repository-name $IMAGE_NAME \
  --region $AWS_REGION

# Build for AMD64 (Linux platform used by ECS)
docker build --platform linux/amd64 -t $IMAGE_NAME:latest -f docker/Dockerfile .

# Tag for ECR
docker tag $IMAGE_NAME:latest \
  $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$IMAGE_NAME:latest

# Push to ECR
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$IMAGE_NAME:latest

# Deploy others
cd scripts
# ./cleanup.sh
./create_secrets.sh
./create_iam_roles.sh
./create_resources.sh
./create_task_definition.sh
cd ..