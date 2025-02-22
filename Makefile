# Project name and paths
APP_NAME := organize-media
BIN_DIR := bin
SRC_DIR := .

# Compilation
build:
	@mkdir -p $(BIN_DIR)                       # Create the bin directory
	go build -o $(BIN_DIR)/$(APP_NAME) $(SRC_DIR)/cmd/main.go

# Cleaning
clean:
	@rm -rf $(BIN_DIR)                         # Delete the bin directory
	@rm -f coverage.out                        # Delete the coverage file
	@rm -rf ./cmd/logs                         # Delete the logs directory
	@rm -rf ./logs 
	@rm -f old.txt new.txt benchmark_results.txt

# Command to run the application
run: build
	./$(BIN_DIR)/$(APP_NAME) --source ../../Pictures/Import/ --dest ../../Pictures/RAW/ --compression 50 --enable-log

# Benchmarking targets
bench: bench-files

bench-files:
	go test -bench=BenchmarkProcessSpecificFiles \
		-count=10 \
		-benchtime=1x \
		-run=^# \
		./pkg/utils/... | tee benchmark_results.txt

.PHONY: build clean run bench bench-files bench-stats deps-benchstat
