all: build install

build:
	go build -o bin/tellme cmd/tellme/main.go

install:
	chmod +x bin/tellme
	ln bin/tellme /usr/local/bin/tellme

