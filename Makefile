# Project name and paths
APP_NAME := organize-pictures
BIN_DIR := bin
SRC_DIR := .

# Compilation
build:
	@mkdir -p $(BIN_DIR)                       # Create the bin directory
	go build -o $(BIN_DIR)/$(APP_NAME) $(SRC_DIR)/main.go

# Cleaning
clean:
	@rm -rf $(BIN_DIR)                         # Delete the bin directory

# Command to run the application
run: build
	./$(BIN_DIR)/$(APP_NAME) /path/to/source /path/to/destination

.PHONY: build clean run