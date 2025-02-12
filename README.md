
# Organize Media: A Tool to Manage and Optimize Your Photo Library

A utility to organize pictures by their taken date. RAW files are moved to designated folders, while JPG files can be optionally compressed before being relocated.

## Codecov Coverage
[![codecov](https://codecov.io/gh/matdmb/organize-media/branch/main/graph/badge.svg?token=4UZGB2L9LB)](https://codecov.io/gh/matdmb/organize-media)

_Current test coverage as tracked by Codecov._

## Features
- Organizes pictures by their taken date.
- Moves RAW files to designated folders.
- Compresses and moves JPG files (optional).
- Lightweight and simple to use.

## Prerequisites
- [Go](https://go.dev/) version `1.19` or later.
- `make` utility (for running commands).

## Installation

```bash
git clone git@github.com:matdmb/organize-media.git
cd organize-media
```

## How to Build the Application

```bash
make build
```

## How to Run the Application

```bash
./bin/organize-media --source <source-folder> --dest <destination-folder> [--compression <compression-level>] [--delete] [--enable-log]
```

- `--source`: Path to the folder containing your pictures.
- `--dest`: Path to the folder where organized pictures will be stored.
- `--compression`: (Optional) Compression level for JPG files (0-100). Defaults to -1 (no compression applied).
- `--delete`: (Optional) Delete source files after processing
- `--enable-log`: (Optional) Save application messages to a log file

Alternatively, use the `make run` command if source and destination folders are set in the `Makefile`.

## Performance analysis

### Benchmark
Measure the performance of some code units with varying sample sizes:

#### Simple benchmark
More suited for quick runs, while developing benchmark tests:
```
make benchmark
```

#### Advanced benchmark
For more reliable and refined data: (Takes about 2 min to process)
```
make deps-benchmark-stats
make benchmark-stats
```

Example output
```
go test -bench=. -count 6 -run=^# -benchtime=0.3s ./... | benchstat -
goos: linux
goarch: amd64
pkg: github.com/matdmb/organize-media/internal
cpu: AMD Ryzen 5 6600U with Radeon Graphics
                              │      -       │
                              │    sec/op    │
ListFiles/with_0_samples-12     45.45µ ±  5%
ListFiles/with_1_samples-12     402.3m ± 29%
ListFiles/with_2_samples-12     615.6m ± 46%
ListFiles/with_5_samples-12     697.0m ±  4%
ListFiles/with_10_samples-12     1.367 ±  6%
ListFiles/with_100_samples-12    24.62 ± 13%
geomean                         253.3m`
```

> With a folder of 100 JPG + 100 RAW, function ListFiles takes about 24 seconds.

### Profile
While benchmark provides measuring of execution time, profiling allows to disassemble the execution.
This can be useful to idenfity bottlenecks.

Example of profiling the ListFiles function:
```
go test -bench=BenchmarkListFiles/with_1_samples -run=^# -benchmem -cpuprofile profile.out ./internal
go tool pprof -http=: profile.out
```

## Cleaning

```bash
make clean
```

## License
This project is licensed under the MIT License. See the `LICENSE` file for more information.

## Contributing
Contributions are welcome! Feel free to submit a pull request or open an issue.
