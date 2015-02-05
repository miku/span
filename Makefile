all:
	goimports -w .
	go build -o span cmd/span/main.go
	go build -o ihparse cmd/ihparse/main.go

clean:
	rm -f span ihparse
