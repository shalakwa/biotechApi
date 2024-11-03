# Biotech API

This project is a command-line tool for processing species data and gene regions using the NCBI Entrez API.

## Prerequisites

- Go 1.16 or later
- Internet connection

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/shalakwa/biotechApi.git
    cd biotechApi
    ```

2. Build the project:
    ```sh
    go build -o biotechApi
    ```
   
or

Copy the built binary from the `build` directory. (might be outdated)


## Usage

### Command-Line Arguments

- `-speciesfile`: Path to the file containing species names.
- `-species`: Species name(Not needed if file is provided).
- `-gene`: Gene region (rbcL or matK) (required, can be set in config).
- `-outputdir`: Directory to store output files.
- `-config`: Configuration file (default: config.json).
- `--setup`: Run initial setup or reconfigure.
- `--help`: Show usage information.

### Examples

1. **Run with a species file and gene region:**
    ```sh
    ./biotechApi -speciesfile="species.txt" -gene="rbcL"
    ```

2. **Run with a single species and gene region:**
    ```sh
    ./biotechApi -species="Homo sapiens" -gene="matK"
    ```

3. **Run setup to create a configuration file:**
    ```sh
    ./biotechApi --setup
    ```

### Configuration

The configuration file (`config.json`) can be created and updated using the `--setup` flag. It contains the following fields:

- `species_file`: Path to the file containing species names.
- `output_dir`: Directory to store output files.
- `gene`: Gene region.

### Cross-Compilation

To cross-compile the project for different operating systems and architectures, use the provided PowerShell script `cross_compile.ps1`:

```sh
.\cross_compile.ps1 <GOOS> <GOARCH>
```

Example:
```sh
.\cross_compile.ps1 linux amd64
```

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.