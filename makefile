.PHONY: test build docker-build run

# 執行所有單元測試
test:
	go test -v ./...

# 編譯執行檔 (靜態連結並優化體積)
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o tinder-server ./cmd/server/main.go

# 建立 Docker Image
docker-build:
	docker build -t tinder-matching-system .

# 快速啟動服務
run:
	go run ./cmd/server/main.go
