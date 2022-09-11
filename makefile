.PHONY=build
build: build/arm5/hashsnap

.PHONY=clean
clean:
	rm build/arm5/hashsnap

build/arm5/hashsnap:
	GOARM=5 GOARCH=arm go build -o $@ cmd/main.go
