package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Config holds configuration settings
type Config struct {
	SpeciesFile string `json:"species_file"`
	OutputDir   string `json:"output_dir"`
	Gene        string `json:"gene"`
}

func main() {
	// Command-line arguments
	speciesFilePtr := flag.String("speciesfile", "", "Path to the file containing species names")
	speciesPtr := flag.String("species", "", "Species name")
	genePtr := flag.String("gene", "", "Gene region (rbcL or matK) (required)")
	outputDirPtr := flag.String("outputdir", "", "Directory to store output files")
	configFilePtr := flag.String("config", "config.json", "Configuration file")
	setupFlag := flag.Bool("setup", false, "Run initial setup or reconfigure")
	helpFlag := flag.Bool("help", false, "Show usage information")
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage:")
		fmt.Println("go run main.go -gene=\"rbcL or matK\" [-species=\"Species Name\"] [-speciesfile=\"species.txt\"] [-outputdir=\"output\"] [--setup]")
		os.Exit(0)
	}

	if *setupFlag {
		err := runSetup()
		if err != nil {
			fmt.Println("Error during setup:", err)
			os.Exit(1)
		}
		fmt.Println("Setup completed successfully.")
		os.Exit(0)
	}

	// Load configuration
	config, err := loadConfig(*configFilePtr)
	if err != nil {
		fmt.Println("Configuration file not found. Please run the script with the '--setup' flag to create one.")
		os.Exit(1)
	}

	// Override config settings with command-line arguments if provided
	if *speciesFilePtr != "" {
		config.SpeciesFile = *speciesFilePtr
	}
	if *outputDirPtr != "" {
		config.OutputDir = *outputDirPtr
	}
	if *genePtr == "" && config.Gene == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Gene region is required. Would you like to update the config to include the gene? (Y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			fmt.Print("Please enter the gene region: ")
			geneInput, _ := reader.ReadString('\n')
			geneInput = strings.TrimSpace(geneInput)

			if geneInput == "" {
				fmt.Println("Gene region cannot be empty.")
				os.Exit(1)
			}

			config.Gene = geneInput
			// Save the updated config if necessary
			saveErr := saveConfig("config.json", config)
			if saveErr != nil {
				return
			}
			fmt.Println("Configuration updated successfully.")
		} else {
			fmt.Println("Gene region is required. Run -setup or use -gene Usage:")
			fmt.Println("./biotechApi -gene=\"rbcL or matK\" [-species=\"Species Name\"] [-speciesfile=\"species.txt\"] [-outputdir=\"output\"] [--setup]")
			os.Exit(1)
		}
	} else if *genePtr != "" {
		config.Gene = *genePtr
	}

	if config.SpeciesFile == "" && *speciesPtr == "" {
		fmt.Println("Either species name or species file must be provided.")
		os.Exit(1)
	}

	gene := *genePtr
	outputDir := config.OutputDir

	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating output directory:", err)
			return
		}
	}

	// Create a rate limiter that allows 2 requests per second
	rateLimiter := time.Tick(time.Second / 2)

	if config.SpeciesFile != "" {
		// Process species from file
		speciesFile := config.SpeciesFile
		// Open the species file
		file, err := os.Open(speciesFile)
		if err != nil {
			fmt.Printf("Error opening species file '%s': %v\n", speciesFile, err)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Printf("Error closing species file '%s': %v\n", speciesFile, err)
			}
		}(file)

		// Read species names line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			species := strings.TrimSpace(scanner.Text())
			if species == "" {
				continue
			}
			fmt.Printf("Processing species: %s\n", species)
			err := processSpecies(species, gene, outputDir, rateLimiter)
			if err != nil {
				fmt.Printf("Error processing species %s: %v\n", species, err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading species file:", err)
			return
		}
	} else if *speciesPtr != "" {
		// Process single species from command line
		species := strings.TrimSpace(*speciesPtr)
		fmt.Printf("Processing species: %s\n", species)
		err := processSpecies(species, gene, outputDir, rateLimiter)
		if err != nil {
			fmt.Printf("Error processing species %s: %v\n", species, err)
		}
	}

	fmt.Println("Processing completed.")
}

func runSetup() error {
	configFile := "config.json"
	config := &Config{}
	reader := bufio.NewReader(os.Stdin)

	// Prompt for species file path
	fmt.Print("Enter path to species file (or leave blank to use command-line species): ")
	speciesFile, _ := reader.ReadString('\n')
	speciesFile = strings.TrimSpace(speciesFile)
	if speciesFile != "" {
		// Check if file exists, create it if not
		if _, err := os.Stat(speciesFile); os.IsNotExist(err) {
			emptyFile, err := os.Create(speciesFile)
			if err != nil {
				fmt.Printf("Error creating species file '%s': %v\n", speciesFile, err)
				return err
			}
			closeErr := emptyFile.Close()
			if closeErr != nil {
				fmt.Printf("Error closing species file '%s': %v\n", speciesFile, closeErr)
				return closeErr
			}
			fmt.Printf("Species file '%s' created. Please add species names to this file.\n", speciesFile)
		}
	}
	config.SpeciesFile = speciesFile

	// Prompt for output directory
	fmt.Print("Enter output directory [default: output]: ")
	outputDir, _ := reader.ReadString('\n')
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		outputDir = "output"
	}
	config.OutputDir = outputDir

	// Prompt for config file path
	fmt.Print("Enter path to save configuration file [default: config.json]: ")
	configFileInput, _ := reader.ReadString('\n')
	configFileInput = strings.TrimSpace(configFileInput)
	if configFileInput != "" {
		configFile = configFileInput
	}

	// Save the config to a file
	err := saveConfig(configFile, config)
	if err != nil {
		return err
	}
	fmt.Println("Configuration saved to", configFile)
	return nil
}

func loadConfig(configFile string) (*Config, error) {

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func saveConfig(filename string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Configuration saved to", filename)
	return nil
}

func processSpecies(species string, gene string, outputDir string, rateLimiter <-chan time.Time) error {
	// Build the query
	db := "nucleotide"

	// Build the search query
	// Exclude complete genomes and unverified sequences
	query := fmt.Sprintf("%s[Organism] AND %s[Gene] NOT complete genome[Title] NOT unverified[Title]", species, gene)

	// Assemble the esearch URL
	base := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/"
	esearchURL := base + "esearch.fcgi"
	esearchParams := url.Values{}
	esearchParams.Set("db", db)
	esearchParams.Set("term", query)
	esearchParams.Set("usehistory", "y")
	esearchParams.Set("retmax", "1000") // Adjust as needed

	// Rate limiting before making the request
	<-rateLimiter

	// Make the esearch request
	resp, err := http.Get(esearchURL + "?" + esearchParams.Encode())
	if err != nil {
		return fmt.Errorf("error making esearch request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading esearch response body: %v", err)
	}
	output := string(bodyBytes)

	// Parse WebEnv and QueryKey
	var webEnv, queryKey string
	webEnvRe := regexp.MustCompile(`<WebEnv>(\S+)</WebEnv>`)
	queryKeyRe := regexp.MustCompile(`<QueryKey>(\d+)</QueryKey>`)

	if matches := webEnvRe.FindStringSubmatch(output); len(matches) > 1 {
		webEnv = matches[1]
	} else {
		return fmt.Errorf("WebEnv not found in esearch response")
	}

	if matches := queryKeyRe.FindStringSubmatch(output); len(matches) > 1 {
		queryKey = matches[1]
	} else {
		return fmt.Errorf("QueryKey not found in esearch response")
	}

	// Check if any results were found
	countRe := regexp.MustCompile(`<Count>(\d+)</Count>`)
	var count string
	if matches := countRe.FindStringSubmatch(output); len(matches) > 1 {
		count = matches[1]
		if count == "0" {
			fmt.Printf("No records found for species '%s' and gene '%s'.\n", species, gene)
			return nil
		}
	} else {
		return fmt.Errorf("Count not found in esearch response")
	}

	// Assemble the efetch URL
	efetchURL := base + "efetch.fcgi"
	efetchParams := url.Values{}
	efetchParams.Set("db", db)
	efetchParams.Set("query_key", queryKey)
	efetchParams.Set("WebEnv", webEnv)
	efetchParams.Set("rettype", "fasta")
	efetchParams.Set("retmode", "text")

	// Rate limiting before making the request
	<-rateLimiter

	// Make the efetch request
	resp, err = http.Get(efetchURL + "?" + efetchParams.Encode())
	if err != nil {
		return fmt.Errorf("error making efetch request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	// Read the response body
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading efetch response body: %v", err)
	}
	data := string(bodyBytes)

	// Check if any FASTA sequences were retrieved
	if len(data) == 0 {
		fmt.Printf("No FASTA sequences found for species '%s' and gene '%s'.\n", species, gene)
		return nil
	}

	// Prepare the output file name
	safeSpeciesName := sanitizeFileName(species)
	outputFileName := fmt.Sprintf("%s_%s.txt", safeSpeciesName, gene)
	outputFilePath := filepath.Join(outputDir, outputFileName)

	// Write the FASTA sequences to the output file
	err = os.WriteFile(outputFilePath, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("error writing to output file '%s': %v", outputFilePath, err)
	}

	fmt.Printf("Results written to '%s'\n", outputFilePath)
	return nil
}

// sanitizeFileName removes or replaces characters that are invalid in file names
func sanitizeFileName(name string) string {
	// Remove any characters that are not letters, numbers, hyphens, or underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
	return re.ReplaceAllString(name, "_")
}
