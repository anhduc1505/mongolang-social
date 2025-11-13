# Set env SERVER_ENV=dev before run -g ./server/engine.go
serve:
	go run main.go serve
swag:
	swag fmt
	swag init --parseDependency --parseDependencyLevel 3 -g main.go -g handler.go -d ./internal/handler -o ./docs/swagger

test:
	go test ./internal/service/... -v

test-coverage:
	go test ./internal/service/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

build-check:
	go build ./internal/...
migrate:
	go run main.go migration migrate --schema --data
rollback:
	go run main.go migration rollback --schema --data