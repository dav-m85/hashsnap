# .PHONY=clean
# clean

# .PHONY=build
# build:
# 	mkdir -p build
build/arm5/hashsnap:
	GOARM=5 GOARCH=arm go build -o $@ main.go
