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
bench: bench-files bench-stats

bench-files:
	go test -bench=BenchmarkProcessSpecificFiles \
		-count=100 \
		-benchtime=1x \
		-run=^# \
		./pkg/utils/... | tee benchmark_results.txt

bench-stats: deps-benchstat
	@echo "Running first benchmark set..."
	go test -bench=BenchmarkProcessSpecificFiles \
		-count=1 \
		-benchmem \
		-benchtime=1x \
		-run=^# \
		./pkg/utils/... > old.txt
	@echo "Running second benchmark set..."
	go test -bench=BenchmarkProcessSpecificFiles \
		-count=1 \
		-benchmem \
		-benchtime=1x \
		-run=^# \
		./pkg/utils/... > new.txt
	@echo "Comparing results..."
	benchstat old.txt new.txt

deps-benchstat:
	go install golang.org/x/perf/cmd/benchstat@latest

.PHONY: build clean run bench bench-files bench-stats deps-benchstat
