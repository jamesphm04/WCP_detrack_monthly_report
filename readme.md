# WCP Detrack Monthly Report

A Go CLI application to fetch monthly job data from **Detrack** and save it to JSON/CSV files.  
Built with Go, structured for Docker, and using **Zap** for logging and **godotenv** for environment variable management.

---

## Features

- Fetches all jobs from Detrack via API.
- Saves jobs to `jobs.json` (and optionally to CSV).
- Logs actions and errors using structured logging (`go.uber.org/zap`).
- Supports configuration via `.env` files.
- Docker-ready for easy deployment.

---

## Prerequisites

- [Go 1.25+](https://golang.org/dl/)
- Docker (optional, if using container)

---

## Installation

Clone the repository:

```bash
git clone https://github.com/jamesphm04/WCP_detrack_monthly_report.git
cd WCP_detrack_monthly_report
```

## Initialize Go modules and install dependencies:

```bash
git mod download
```

## Create your `.env`

```bash
# Detrack API Configuration
BASE_URL=https://app.detrack.com/api/v2
API_KEY=<your_api_key_here>

# Email Notification
EMAIL_SENDER=<your_email_here>
EMAIL_PASSWORD=<your_email_app_pwd_here>
EMAIL_RECEIVERS=<your_comma_separated_emails_year>
```

## Running Locally

```bash
go run ./cmd/main.go
```

## Running with Docker

```bash
docker build -t wcp-detrack:latest .

docker run -d `
  --name detrack `
  --restart unless-stopped `
  --env-file .env `
  --memory 256m `
  --cpus 1 `
  --read-only `
  --cap-drop ALL `
  wcp-detrack:latest

```

## Troubleshooting

- When your run docker, there is might be an error with DNS, just run that again.

## DEPLOY

export IAM_USER=$(aws iam get-user --query "User.UserName" --output text)

aws iam attach-user-policy \
 --user-name $IAM_USER \
 --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser
