SOURCES=config.go \
	main.go

all: daikinweb

daikinweb: ${SOURCES}
	go get
	go build

run: daikinweb
	./daikinweb

lint: main.go
	go fmt

clean:
	rm daikinweb
