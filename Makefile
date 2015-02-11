all:
	goimports -w .
	go build -o span cmd/span/main.go

clean:
	rm -f span

test:
	go test -cover ./...
