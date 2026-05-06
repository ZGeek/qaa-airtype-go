.PHONY: build clean run

build:
	cd go && go build -ldflags="-H windowsgui" -o ../QAA-AirType-Go.exe ./cmd/airtype

clean:
	rm -f QAA-AirType-Go.exe

run:
	cd go && go run ./cmd/airtype