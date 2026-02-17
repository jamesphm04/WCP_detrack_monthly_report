#!/bin/bash
set -e

# -------------------------------
# Variables
# -------------------------------
AWS_REGION=ap-southeast-2
AWS_ACCOUNT_ID=380206744475
TASK_NAME=wcp-detrack-monthly-report
IMAGE_NAME=wcp-detrack-monthly-report
SG_NAME=wcp-detrack-monthly-report-sg
LOG_GROUP_NAME=wcp-detrack-monthly-report

# Secret names
SECRET_API_KEY=wcp-detrack-monthly-report-api-key
SECRET_EMAIL_SENDER=wcp-detrack-monthly-report-email-sender
SECRET_EMAIL_PASSWORD=wcp-detrack-monthly-report-email-password
SECRET_EMAIL_RECEIVERS=wcp-detrack-monthly-report-email-receivers

# -------------------------------
# Delete ECS Task Definition
# -------------------------------
echo "Deleting ECS task definitions for ${TASK_NAME}..."
aws ecs list-task-definitions --family-prefix $TASK_NAME --region $AWS_REGION --query "taskDefinitionArns" --output text | \
  xargs -n 1 aws ecs deregister-task-definition --region $AWS_REGION --task-definition

# -------------------------------
# Delete ECS Cluster
# -------------------------------
echo "Deleting ECS cluster 'default'..."
aws ecs delete-cluster --cluster default --region $AWS_REGION || true

# -------------------------------
# Delete Security Group
# -------------------------------
SG_ID=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=${SG_NAME}" --query "SecurityGroups[0].GroupId" --output text --region $AWS_REGION)
if [ "$SG_ID" != "None" ]; then
  echo "Deleting Security Group $SG_NAME..."
  aws ec2 delete-security-group --group-id $SG_ID --region $AWS_REGION
fi

# -------------------------------
# Delete CloudWatch Log Group
# -------------------------------
echo "Deleting CloudWatch log group /ecs/${LOG_GROUP_NAME}..."
aws logs delete-log-group --log-group-name /ecs/${LOG_GROUP_NAME} --region $AWS_REGION || true

# -------------------------------
# Delete IAM Roles and Policies
# -------------------------------
echo "Deleting IAM policies and roles..."

# ECS Execution Role
aws iam delete-role-policy --role-name ecsTaskExecutionRole --policy-name SecretsAccess || true
aws iam detach-role-policy --role-name ecsTaskExecutionRole --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy || true
aws iam delete-role --role-name ecsTaskExecutionRole || true

# EventBridge Role
aws iam delete-role-policy --role-name ecsEventsRole --policy-name ECSRunTask || true
aws iam delete-role --role-name ecsEventsRole || true

# -------------------------------
# Delete Secrets
# -------------------------------
echo "Deleting Secrets..."
for secret in $SECRET_API_KEY $SECRET_EMAIL_SENDER $SECRET_EMAIL_PASSWORD $SECRET_EMAIL_RECEIVERS; do
  aws secretsmanager delete-secret --secret-id $secret --region $AWS_REGION --force-delete-without-recovery || true
done

echo "✅ Cleanup complete!"
