# Tool for managing picture files
A utility to organize pictures by their taken date. RAW files are moved to designated folders, while JPG files can be optionally compressed before being relocated.

## Test coverage insights
[![codecov](https://codecov.io/gh/matdmb/organize_pictures/graph/badge.svg?token=4UZGB2L9LB)](https://codecov.io/gh/matdmb/organize_pictures)

## Codecov Coverage Graph
![Codecov Sunburst Graph](https://codecov.io/gh/matdmb/organize_pictures/graphs/sunburst.svg?token=4UZGB2L9LB)

## Installation

```bash
git clone git@github.com:matdmb/organize_pictures.git
cd organize_pictures
```

## How to build the application

```bash
make build
```

## How to run the application

```bash
./bin/organize-pictures ../src/ ../dest compression
```
You can also use the "make run" command if source and destination folders have been correctly set up in the Makefile

## Cleaning

```bash
make clean
```