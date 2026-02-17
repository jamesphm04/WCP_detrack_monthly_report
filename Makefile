build:
	go build -o bin/main ./cmd/main.go

run: build
	./bin/main

deploy:
	./scripts/deploy.sh

run_task:
	./scripts/run_task.sh