test:
	go test -cover ./...

cover:
	govendor test -v -coverprofile=cover.out +local
	go tool cover -html=cover.out
	rm cover.out

fetch:
	govendor fetch -v +outside

sync:
	govendor sync -v
