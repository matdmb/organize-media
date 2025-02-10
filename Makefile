# Project name and paths
APP_NAME := organize-media
BIN_DIR := bin
SRC_DIR := .


# Benchmark
benchmark:
	go test -bench=. -count 1 -run=^# -benchtime=0.3s ./... -v

deps-benchmark-stats:
	go install golang.org/x/perf/cmd/benchstat

benchmark-stats:
	go test -bench=. -count 6 -run=^# -benchtime=0.3s ./... | benchstat -

# Compilation
build:
	@mkdir -p $(BIN_DIR)                       # Create the bin directory
	go build -o $(BIN_DIR)/$(APP_NAME) $(SRC_DIR)/main.go

# Cleaning
clean:
	@rm -rf $(BIN_DIR)                         # Delete the bin directory
	@rm -f coverage.out                        # Delete the coverage file

# Command to run the application
run: build
	./$(BIN_DIR)/$(APP_NAME) --source ../../Pictures/Import/ --dest ../../Pictures/RAW/ --compression 50

.PHONY: build clean run
