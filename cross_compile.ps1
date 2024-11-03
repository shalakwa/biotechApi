# cross_compile.ps1

$ErrorActionPreference = "Stop"

$PROJECT_NAME = "biotechApi"
$OUTPUT_DIR = "build"

# Save current GOOS and GOARCH environment variables
$OLD_GOOS = $env:GOOS
$OLD_GOARCH = $env:GOARCH

# Get GOOS and GOARCH from command-line arguments or environment variables
if ($args.Count -eq 2) {
    $GOOS = $args[0]
    $GOARCH = $args[1]
} elseif ($env:GOOS -and $env:GOARCH) {
    Write-Host "Using GOOS=$($env:GOOS) and GOARCH=$($env:GOARCH) from environment variables"
    $GOOS = $env:GOOS
    $GOARCH = $env:GOARCH
} else {
    Write-Host "Usage: .\cross_compile.ps1 <GOOS> <GOARCH>"
    exit 1
}

# Set the output file name
$OUTPUT_NAME = "$PROJECT_NAME-$GOOS-$GOARCH"
if ($GOOS -eq "windows") {
    $OUTPUT_NAME += ".exe"
}

# Create the output directory if it doesn't exist
if (!(Test-Path -Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

# Set the GOOS and GOARCH environment variables for the build
$env:GOOS = $GOOS
$env:GOARCH = $GOARCH

# Build the project
Write-Host "Building for $GOOS/$GOARCH..."
go build -o "$OUTPUT_DIR\$OUTPUT_NAME"

Write-Host "Build completed successfully."

# Restore the original GOOS and GOARCH environment variables
$env:GOOS = $OLD_GOOS
$env:GOARCH = $OLD_GOARCH
