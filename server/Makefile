.PHONY: build run-sim run-ingest clean

build:
	go build -o bin/spanner-sim ./cmd/sim
	go build -o bin/spanner-ingest ./cmd/ingest

run-sim:
	./bin/spanner-sim

run-ingest:
	./bin/spanner-ingest

clean:
	rm -rf bin/
