docker-build:
	docker-compose build
docker-run:
	docker-compose up -d
build: 
	go build -a -o dns-proxy cmd/main.go
run: 
	source ./env.env && go run ./cmd/main.go
