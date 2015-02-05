all:
	goimports -w .
	go build -o crparse cmd/crparse/main.go
	go build -o ihparse cmd/ihparse/main.go
	go build -o span cmd/span/main.go

clean:
	rm -f span ihparse crparse
