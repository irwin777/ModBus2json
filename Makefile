build:
	GOARCH=arm GOARM=7 go build -o bin/test main.go
install: build
	scp bin/test root@192.168.88.225:~/
