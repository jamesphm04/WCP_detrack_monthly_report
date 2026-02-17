#!/bin/bash

export AWS_REGION=ap-southeast-2
export AWS_ACCOUNT_ID=380206744475

# Get VPC ID
export VPC_ID=$(aws ec2 describe-vpcs \
  --filters "Name=isDefault,Values=true" \
  --query "Vpcs[0].VpcId" \
  --output text \
  --region $AWS_REGION)

# Get Subnet ID
export SUBNET_ID=$(aws ec2 describe-subnets \
  --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "Subnets[0].SubnetId" \
  --output text \
  --region $AWS_REGION)

# Get Security Group ID
export SG_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=wcp-detrack-monthly-report-sg" \
  --query "SecurityGroups[0].GroupId" \
  --output text \
  --region $AWS_REGION)

echo "VPC_ID: $VPC_ID"
echo "SUBNET_ID: $SUBNET_ID"
echo "SG_ID: $SG_ID"

aws ecs run-task \
  --cluster default \
  --task-definition wcp-detrack-monthly-report \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNET_ID],securityGroups=[$SG_ID],assignPublicIp=ENABLED}" \
  --region $AWS_REGION