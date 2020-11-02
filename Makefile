all: test

models:
	@flatc --go --gen-mutable --gen-object-api --gen-onefile --go-namespace errx ./models.fbs

test: models
	@goimports -w .
	@go test -timeout 10s -race -count 10 -cover -coverprofile=./errx.cover ./...

cover: test
	@go tool cover -html=./errx.cover
