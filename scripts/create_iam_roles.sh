#!/bin/bash

# -------------------------------
# Variables
# -------------------------------
export AWS_REGION=ap-southeast-2
export AWS_ACCOUNT_ID=380206744475
export TASK_NAME=wcp-detrack-monthly-report

# Secret names (each key has its own secret)
export SECRET_API_KEY=wcp-detrack-monthly-report-api-key
export SECRET_EMAIL_SENDER=wcp-detrack-monthly-report-email-sender
export SECRET_EMAIL_PASSWORD=wcp-detrack-monthly-report-email-password
export SECRET_EMAIL_RECEIVERS=wcp-detrack-monthly-report-email-receivers

# -------------------------------
# ECS Execution Role
# -------------------------------
cat > ../infra/ecs-trust.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "ecs-tasks.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOF

# Create role if not exists
aws iam get-role --role-name ecsTaskExecutionRole >/dev/null 2>&1 || \
aws iam create-role \
  --role-name ecsTaskExecutionRole \
  --assume-role-policy-document file://../infra/ecs-trust.json

aws iam attach-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy

# -------------------------------
# Add Secrets Access Policy
# -------------------------------
cat > ../infra/secrets-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "kms:Decrypt"
    ],
    "Resource": [
      "arn:aws:secretsmanager:${AWS_REGION}:${AWS_ACCOUNT_ID}:secret:${SECRET_API_KEY}*",
      "arn:aws:secretsmanager:${AWS_REGION}:${AWS_ACCOUNT_ID}:secret:${SECRET_EMAIL_SENDER}*",
      "arn:aws:secretsmanager:${AWS_REGION}:${AWS_ACCOUNT_ID}:secret:${SECRET_EMAIL_PASSWORD}*",
      "arn:aws:secretsmanager:${AWS_REGION}:${AWS_ACCOUNT_ID}:secret:${SECRET_EMAIL_RECEIVERS}*"
    ]
  }]
}
EOF

aws iam put-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-name SecretsAccess \
  --policy-document file://../infra/secrets-policy.json

# -------------------------------
# EventBridge Role
# -------------------------------
cat > ../infra/events-trust.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "events.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOF

# Create role if not exists
aws iam get-role --role-name ecsEventsRole >/dev/null 2>&1 || \
aws iam create-role \
  --role-name ecsEventsRole \
  --assume-role-policy-document file://../infra/events-trust.json

cat > ../infra/events-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ecs:RunTask",
      "Resource": "arn:aws:ecs:${AWS_REGION}:${AWS_ACCOUNT_ID}:task-definition/${TASK_NAME}:*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/ecsTaskExecutionRole"
    }
  ]
}
EOF

aws iam put-role-policy \
  --role-name ecsEventsRole \
  --policy-name ECSRunTask \
  --policy-document file://../infra/events-policy.json
