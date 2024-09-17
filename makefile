install:
	go build -o go-for-milk main.go
	cp ./go-for-milk $(GOPATH)/bin/rtm
	rm ./go-for-milk

