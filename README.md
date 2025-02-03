
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
./bin/organize-media --source <source-folder> --dest <destination-folder> [--compression <compression-level>] [--delete <true|false>]
```

- `-source`: Path to the folder containing your pictures.
- `-dest`: Path to the folder where organized pictures will be stored.
- `-compression`: (Optional) Compression level for JPG files (0-100). Defaults to -1 (no compression applied).
- `-delete`: (Optional) Delete source files after processing

Alternatively, use the `make run` command if source and destination folders are set in the `Makefile`.

## Cleaning

```bash
make clean
```

## License
This project is licensed under the MIT License. See the `LICENSE` file for more information.

## Contributing
Contributions are welcome! Feel free to submit a pull request or open an issue.
