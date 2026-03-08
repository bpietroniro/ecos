.PHONY: build run test vet lint clean dynamo-local dynamo-setup docker-build

APP_NAME := ingestion
PORT := 8080

build:
	go build -o bin/$(APP_NAME) ./cmd/server

run: build
	DYNAMODB_ENDPOINT=http://localhost:8000 \
	PORT=$(PORT) \
	./bin/$(APP_NAME)

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@echo "Lint passed (go vet)"

clean:
	rm -rf bin/

dynamo-local:
	docker run -d --name ecos-dynamodb \
		-p 8000:8000 \
		amazon/dynamodb-local:latest

dynamo-setup:
	aws dynamodb create-table \
		--endpoint-url http://localhost:8000 \
		--table-name ecos-sensor-readings \
		--attribute-definitions \
			AttributeName=stationId,AttributeType=S \
			AttributeName=sk,AttributeType=S \
			AttributeName=readingType,AttributeType=S \
		--key-schema \
			AttributeName=stationId,KeyType=HASH \
			AttributeName=sk,KeyType=RANGE \
		--global-secondary-indexes \
			'[{"IndexName":"byReadingType","KeySchema":[{"AttributeName":"readingType","KeyType":"HASH"},{"AttributeName":"sk","KeyType":"RANGE"}],"Projection":{"ProjectionType":"ALL"}}]' \
		--billing-mode PAY_PER_REQUEST \
		--no-cli-pager
	aws dynamodb create-table \
		--endpoint-url http://localhost:8000 \
		--table-name ecos-stations \
		--attribute-definitions \
			AttributeName=id,AttributeType=S \
		--key-schema \
			AttributeName=id,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST \
		--no-cli-pager

docker-build:
	docker build -t ecos-$(APP_NAME) .
