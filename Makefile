

all: daikinweb

daikinweb: main.go config.go
	go get
	go build

run: daikinweb
	./daikinweb

lint: main.go
	go fmt

clean:
	rm daikinweb
