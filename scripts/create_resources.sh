#!/bin/bash

# Set variables
export AWS_REGION=ap-southeast-2
export AWS_ACCOUNT_ID=380206744475
export SG_NAME=wcp-detrack-monthly-report-sg
export LOG_GROUP_NAME=wcp-detrack-monthly-report

# CloudWatch Log Group
aws logs create-log-group \
  --log-group-name /ecs/${LOG_GROUP_NAME} \
  --region $AWS_REGION

# ECS Cluster
aws ecs create-cluster \
  --cluster-name default \
  --region $AWS_REGION

# Get VPC/Subnet
export VPC_ID=$(aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" --query "Vpcs[0].VpcId" --output text --region $AWS_REGION)
export SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[0].SubnetId" --output text --region $AWS_REGION)

# Security Group
export SG_ID=$(aws ec2 create-security-group \
  --group-name ${SG_NAME} \
  --description "WCP Detrack Monthly Report SG" \
  --vpc-id $VPC_ID \
  --region $AWS_REGION \
  --query 'GroupId' \
  --output text)

aws ec2 authorize-security-group-egress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0 \
  --region $AWS_REGION

aws ec2 authorize-security-group-egress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 587 \
  --cidr 0.0.0.0/0 \
  --region $AWS_REGION