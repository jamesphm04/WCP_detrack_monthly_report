#!/bin/bash

export AWS_REGION=ap-southeast-2
export AWS_ACCOUNT_ID=380206744475
export IMAGE_NAME=wcp-detrack-monthly-report

# Get full secret ARNs
export SECRET_API_KEY_ARN=$(aws secretsmanager describe-secret \
  --secret-id wcp-detrack-monthly-report-api-key \
  --region $AWS_REGION \
  --query 'ARN' \
  --output text)

export SECRET_EMAIL_SENDER_ARN=$(aws secretsmanager describe-secret \
  --secret-id wcp-detrack-monthly-report-email-sender \
  --region $AWS_REGION \
  --query 'ARN' \
  --output text)

export SECRET_EMAIL_PASSWORD_ARN=$(aws secretsmanager describe-secret \
  --secret-id wcp-detrack-monthly-report-email-password \
  --region $AWS_REGION \
  --query 'ARN' \
  --output text)

export SECRET_EMAIL_RECEIVERS_ARN=$(aws secretsmanager describe-secret \
  --secret-id wcp-detrack-monthly-report-email-receivers \
  --region $AWS_REGION \
  --query 'ARN' \
  --output text)

cat > ../infra/task-definition.json <<EOF
{
  "family": "wcp-detrack-monthly-report",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "${IMAGE_NAME}",
      "image": "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${IMAGE_NAME}:latest",
      "essential": true,
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/wcp-detrack-monthly-report",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "environment": [
        {"name": "BASE_URL", "value": "https://app.detrack.com/api/v2"},
        {"name": "FETCH_LIMIT", "value": "1000"},
        {"name": "SMTP_HOST", "value": "smtp.gmail.com"},
        {"name": "SMTP_PORT", "value": "587"}
      ],
      "secrets": [
        {"name": "API_KEY", "valueFrom": "${SECRET_API_KEY_ARN}"},
        {"name": "EMAIL_SENDER", "valueFrom": "${SECRET_EMAIL_SENDER_ARN}"},
        {"name": "EMAIL_PASSWORD", "valueFrom": "${SECRET_EMAIL_PASSWORD_ARN}"},
        {"name": "EMAIL_RECEIVERS", "valueFrom": "${SECRET_EMAIL_RECEIVERS_ARN}"}
      ]
    }
  ]
}
EOF

echo "Task definition created. Registering..."

aws ecs register-task-definition \
  --cli-input-json file://../infra/task-definition.json \
  --region $AWS_REGION

echo "Task definition registered successfully!"