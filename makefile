.PHONY=build
build: build/arm5/hsnap

.PHONY=clean
clean:
	rm build/arm5/hsnap

build/arm5/hsnap:
	GOARM=5 GOARCH=arm go build -o $@ main.go
