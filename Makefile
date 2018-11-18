test:
	go test -cover ./...

cover:
	go test -v -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	rm cover.out
