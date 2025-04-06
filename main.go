package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anchore/go-collections"
	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
	"github.com/anchore/syft/syft/format/cyclonedxjson"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/anchore/syft/syft/source/sourceproviders"
	"github.com/gorilla/mux"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Constants for better maintainability
const (
	contentTypeJSON       = "application/json"
	contentTypeTextPlain  = "text/plain"
	defaultSBOMOutputFile = "sbom.cyclonedx.json"
	defaultLogFile        = "static/output.log"
	defaultLlamaIndexHost = "http://llama-index-api:8000"
	defaultOllamaHost     = "http://host.docker.internal:11434"
	defaultModel          = "mistral"
	gitCloneDir           = "/tmp/git-sbom"
)

// Configuration struct for application settings
type Config struct {
	LlamaIndexEndpoint string
	OllamaHost         string
	DefaultModel       string
	LogFile            string
	SBOMOutputFile     string
}

// Global configuration with defaults
var appConfig = Config{
	LlamaIndexEndpoint: getEnv("LLAMA_INDEX_ENDPOINT", defaultLlamaIndexHost),
	OllamaHost:         getEnv("OLLAMA_HOST", defaultOllamaHost),
	DefaultModel:       getEnv("DEFAULT_MODEL", defaultModel),
	LogFile:            defaultLogFile,
	SBOMOutputFile:     defaultSBOMOutputFile,
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LlamaIndexClient handles interactions with the LlamaIndex API
type LlamaIndexClient struct {
	BaseURL string
}

// NewLlamaIndexClient creates a new client for interacting with LlamaIndex
func NewLlamaIndexClient(baseURL string) *LlamaIndexClient {
	return &LlamaIndexClient{
		BaseURL: baseURL,
	}
}

// AnalyzeVulnerabilities sends vulnerability data to LlamaIndex for enhanced analysis
func (c *LlamaIndexClient) AnalyzeVulnerabilities(scanResults string, sbomData string) (string, error) {
	payload := map[string]interface{}{
		"query": "Analyze these vulnerabilities and provide a comprehensive remediation plan",
		"data": map[string]string{
			"scan_results": scanResults,
			"sbom_data":    sbomData,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/query", contentTypeJSON, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to call LlamaIndex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LlamaIndex API returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse LlamaIndex response: %w", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", errors.New("invalid response format from LlamaIndex")
	}

	return response, nil
}

// Logger provides structured logging
type Logger struct {
	file   *os.File
	logger *log.Logger
}

// NewLogger creates a new structured logger
func NewLogger(filepath string) (*Logger, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags),
	}, nil
}

// Log writes a log entry with timestamp
func (l *Logger) Log(message string) {
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("%s %s", currentTime, message)
	l.logger.Println(logMessage)
}

// Close closes the log file
func (l *Logger) Close() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			l.logger.Printf("Error closing log file: %v", err)
		}
	}
}

// Extract script block from text
func extractScriptBlock(text string) string {
	start := strings.Index(text, "```bash")
	if start == -1 {
		start = strings.Index(text, "```")
	}
	if start == -1 {
		return ""
	}
	rest := text[start+3:]
	end := strings.Index(rest, "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(rest[5:end])
}

// Global logger
var logger *Logger

func main() {
	var err error
	logger, err = NewLogger(appConfig.LogFile)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	r := mux.NewRouter()

	// Add CORS middleware
	r.Use(corsMiddleware)

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static"))).Methods("GET")
	fmt.Println("Serving static files from ./static")

	// API routes
	r.HandleFunc("/generate-sbom", generateSBOMHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/scan-sbom", scanSBOMHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/logs", logsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/remediate", remediateHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/llamaindex-analyze", llamaIndexAnalyzeHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", healthCheckHandler).Methods("GET", "OPTIONS")

	port := getEnv("PORT", "3000")
	fmt.Printf("API is running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":        "healthy",
		"llamaIndexAPI": appConfig.LlamaIndexEndpoint,
		"ollamaHost":    appConfig.OllamaHost,
		"version":       "1.0.0",
	})
}

func generateSBOMHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SBOMSource string `json:"sbomSource"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	source := body.SBOMSource

	if source == "" {
		msg := "Error: No valid source provided. Provide an image, directory path, or remote URL."
		logger.Log(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	sourceInput, err := determineSourceInput(source)
	if err != nil {
		logger.Log(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log(fmt.Sprintf("Processing SBOM for source: %s", sourceInput))

	schemeSource, newUserInput := stereoscope.ExtractSchemeSource(sourceInput, allSourceTags()...)
	getSourceCfg := syft.DefaultGetSourceConfig()
	if schemeSource != "" {
		getSourceCfg = getSourceCfg.WithSources(schemeSource)
		sourceInput = newUserInput
	}

	src, err := syft.GetSource(context.Background(), sourceInput, getSourceCfg)
	if err != nil {
		msg := fmt.Sprintf("Failed to get source: %v", err)
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	cfg := syft.DefaultCreateSBOMConfig().WithCatalogerSelection(
		pkgcataloging.NewSelectionRequest().WithDefaults(
			pkgcataloging.InstalledTag,
			pkgcataloging.DirectoryTag,
			pkgcataloging.ImageTag,
		),
	)
	sbomData, err := syft.CreateSBOM(context.Background(), src, cfg)
	if err != nil {
		msg := fmt.Sprintf("Failed to create SBOM: %v", err)
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// Save SBOM to file
	saveErr := saveSBOMToFile(sbomData, appConfig.SBOMOutputFile)
	if saveErr != nil {
		logger.Log(fmt.Sprintf("Failed to save SBOM to file: %v", err))
		http.Error(w, "Failed to save SBOM to file", http.StatusInternalServerError)
		return
	}

	// Read SBOM content for response
	sbomContent, err := os.ReadFile(appConfig.SBOMOutputFile)
	if err != nil {
		msg := "Failed to read generated SBOM file."
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	logger.Log("SBOM generated successfully.")

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "SBOM generated successfully",
		"format":   "CycloneDX JSON",
		"file":     appConfig.SBOMOutputFile,
		"sbomData": string(sbomContent),
	})
}

func determineSourceInput(source string) (string, error) {
	if _, err := os.Stat(source); err == nil {
		return "dir:" + source, nil
	} else if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		if err := cloneGitRepo(source, gitCloneDir); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w", err)
		}
		return "dir:" + gitCloneDir, nil
	} else {
		return "image:" + source, nil
	}
}

// getQualityScore calculates the quality score of an SBOM using the sbomqs tool
func getQualityScore(sbomFile string) (map[string]interface{}, error) {
	logger.Log(fmt.Sprintf("Calculating SBOM quality score for: %s", sbomFile))

	// Check if sbomqs is installed
	_, err := exec.LookPath("sbomqs")
	if err != nil {
		logger.Log(fmt.Sprintf("sbomqs tool not installed: %v", err))
		return map[string]interface{}{
			"error":                     "sbomqs tool not installed. Please install it to enable quality scoring.",
			"score":                     0.0,
			"installation_instructions": "Install sbomqs using: wget -q https://github.com/interlynk-io/sbomqs/releases/download/v1.0.3/sbomqs-linux-amd64.tar.gz && tar -xzf sbomqs-linux-amd64.tar.gz && sudo mv sbomqs-linux-amd64/sbomqs /usr/local/bin/ && sudo chmod +x /usr/local/bin/sbomqs",
			"documentation_url":         "https://github.com/interlynk-io/sbomqs",
		}, nil // Return info about missing tool but don't fail the whole scan
	}

	// Run sbomqs command with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sbomqs", "score", sbomFile, "--format", "json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Try to extract basic score if available
		basicCmd := exec.Command("sbomqs", "score", sbomFile)
		basicOutput, basicErr := basicCmd.CombinedOutput()

		if basicErr == nil {
			// Parse the basic output for score (format: "7.6     samples/sbom.spdx.yaml")
			scoreRegex := regexp.MustCompile(`([0-9.]+)\s+`)
			matches := scoreRegex.FindStringSubmatch(string(basicOutput))
			if len(matches) > 1 {
				score, parseErr := strconv.ParseFloat(matches[1], 64)
				if parseErr == nil {
					return map[string]interface{}{
						"score":   score,
						"details": "Basic score only. Full details not available.",
					}, nil
				}
			}
		}

		logger.Log(fmt.Sprintf("Failed to calculate SBOM quality score: %v", err))
		return map[string]interface{}{
			"error":      fmt.Sprintf("Failed to calculate quality score: %v", err),
			"score":      0.0,
			"raw_output": string(output),
		}, nil // Don't fail the scan if quality score calculation fails
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		logger.Log(fmt.Sprintf("Failed to parse sbomqs output: %v", err))
		return map[string]interface{}{
			"error":      fmt.Sprintf("Failed to parse quality score output: %v", err),
			"score":      0.0,
			"raw_output": string(output),
		}, nil // Don't fail the scan if quality score parsing fails
	}

	return result, nil
}

func scanSBOMHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SBOMFile    string `json:"sbomFile"`
		UseAdvanced bool   `json:"useAdvanced"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sbomFile := body.SBOMFile
	if sbomFile == "" {
		sbomFile = appConfig.SBOMOutputFile
	}

	if _, err := os.Stat(sbomFile); os.IsNotExist(err) {
		http.Error(w, "SBOM file not found. Please generate it first.", http.StatusBadRequest)
		return
	}

	logger.Log("Starting SBOM scan...")

	scanOutput, err := runGrypeScan(sbomFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Error running Grype: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(scanOutput) == 0 {
		http.Error(w, "No vulnerabilities found", http.StatusOK)
		return
	}

	// Extract SBOM content for advanced analysis
	sbomContent, err := os.ReadFile(appConfig.SBOMOutputFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading SBOM file: %v", err))
		http.Error(w, fmt.Sprintf("Error reading SBOM file: %v", err), http.StatusInternalServerError)
		return
	}

	pkgType := detectPackageType(scanOutput)
	logger.Log(fmt.Sprintf("Detected package type: %s", pkgType))

	// Calculate SBOM quality score
	qualityScore, scoreErr := getQualityScore(sbomFile)
	if scoreErr != nil {
		logger.Log(fmt.Sprintf("Warning: Error calculating quality score: %v", scoreErr))
		// Continue with scan - quality score is optional
	}

	remediation, err := getRemediation(scanOutput, pkgType, body.UseAdvanced, string(sbomContent))
	if err != nil {
		// Don't fail completely, just log the error and proceed with basic scan results
		logger.Log(fmt.Sprintf("Warning: Could not get remediation script: %v", err))
		remediationError := fmt.Sprintf("Could not generate remediation script: %v", err)

		// Still return the scan results without remediation
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"scanResult":         scanOutput,
			"remediationScript":  "",
			"remediationWarning": remediationError,
			"qualityScore":       qualityScore,
		})
		return
	}

	logger.Log("SBOM scan completed successfully.")

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":             "Grype scan and remediation completed successfully",
		"scanResult":          scanOutput,
		"remediationScript":   remediation,
		"remediationCommands": extractScriptBlock(remediation),
		"pkgType":             pkgType,
		"markdownResponse":    fmt.Sprintf("```bash\n%s\n```", extractScriptBlock(remediation)),
		"ollamaModel":         appConfig.DefaultModel,
		"ollamaRawResponse":   remediation,
		"usedLlamaIndex":      body.UseAdvanced,
		"qualityScore":        qualityScore,
	})
}

func runGrypeScan(sbomFile string) (string, error) {
	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running Grype: %w", err)
	}
	return string(output), nil
}

func getRemediation(scanOutput string, pkgType string, useAdvanced bool, sbomContent string) (string, error) {
	if len(scanOutput) == 0 {
		return "", nil // No vulnerabilities, no need for remediation
	}

	// Try advanced analysis if requested
	if useAdvanced {
		// Try to use LlamaIndex for advanced analysis
		client := NewLlamaIndexClient(appConfig.LlamaIndexEndpoint)
		if client != nil {
			// Advanced client creation successful
			llamaResponse, err := client.AnalyzeVulnerabilities(scanOutput, sbomContent)
			if err == nil {
				// Advanced analysis successful
				return llamaResponse, nil
			}
			// Log error but continue to basic remediation
			logger.Log(fmt.Sprintf("advanced analysis failed, falling back to basic: %v", err))
		} else {
			// Log error but continue to basic remediation
			logger.Log("could not initialize advanced analysis client, falling back to basic")
		}
	}

	// Basic remediation with Ollama
	ollamaErr := checkOllamaAvailability()
	if ollamaErr != nil {
		// Generate a simple remediation based on scan output if Ollama is not available
		return generateBasicRemediation(scanOutput, pkgType), nil
	}

	return getOllamaRemediation(scanOutput, pkgType)
}

// generateBasicRemediation creates a simple remediation script based on the scan output
// without requiring Ollama or other external services
func generateBasicRemediation(_ string, pkgType string) string {
	// Extract vulnerability information from scan output
	var recommendations []string

	// Simple parsing of the scan output to identify packages and versions
	var packageManager string
	var updateCommand string

	switch pkgType {
	case "python":
		packageManager = "pip"
		updateCommand = "pip install --upgrade"
	case "nodejs":
		packageManager = "npm"
		updateCommand = "npm update"
	case "java":
		packageManager = "maven"
		updateCommand = "mvn versions:use-latest-versions"
	case "golang":
		packageManager = "go"
		updateCommand = "go get -u"
	case "ruby":
		packageManager = "gem"
		updateCommand = "gem update"
	case "rust":
		packageManager = "cargo"
		updateCommand = "cargo update"
	default:
		packageManager = "unknown"
		updateCommand = "# Unknown package manager"
	}

	// Add basic header with instructions
	recommendations = append(recommendations, fmt.Sprintf("# Basic remediation script for %s packages", packageManager))
	recommendations = append(recommendations, "# Please review before executing")
	recommendations = append(recommendations, "")

	// Add update all command as a fallback
	recommendations = append(recommendations, fmt.Sprintf("# Option 1: Update all packages"))
	recommendations = append(recommendations, fmt.Sprintf("%s", updateCommand))
	recommendations = append(recommendations, "")

	// Add general security advice
	recommendations = append(recommendations, "# Additional security recommendations:")
	recommendations = append(recommendations, "# 1. Review your dependencies and remove unused ones")
	recommendations = append(recommendations, "# 2. Consider using a dependency lockfile")
	recommendations = append(recommendations, "# 3. Set up automated vulnerability scanning in your CI/CD pipeline")

	return strings.Join(recommendations, "\n")
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(appConfig.LogFile)
	if err != nil {
		http.Error(w, "Failed to read logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeTextPlain)
	w.Write(content)
}

// New handler for LlamaIndex direct analysis
func llamaIndexAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query    string `json:"query"`
		ScanData string `json:"scanData"`
		SBOMFile string `json:"sbomFile"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sbomFile := body.SBOMFile
	if sbomFile == "" {
		sbomFile = appConfig.SBOMOutputFile
	}

	if _, err := os.Stat(sbomFile); os.IsNotExist(err) {
		http.Error(w, "SBOM file not found. Please generate it first.", http.StatusBadRequest)
		return
	}

	sbomContent, err := os.ReadFile(sbomFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to read SBOM file: %v", err))
		http.Error(w, "Failed to read SBOM file", http.StatusInternalServerError)
		return
	}

	// Use LlamaIndex for analysis
	client := NewLlamaIndexClient(appConfig.LlamaIndexEndpoint)
	query := body.Query
	if query == "" {
		query = "Analyze these vulnerabilities and provide a comprehensive remediation plan"
	}

	scanData := body.ScanData
	if scanData == "" {
		// Run Grype to get scan data
		scanData, err = runGrypeScan(sbomFile)
		if err != nil {
			logger.Log(fmt.Sprintf("Error running Grype: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	logger.Log("Running LlamaIndex analysis...")
	llamaResponse, err := client.AnalyzeVulnerabilities(scanData, string(sbomContent))
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get LlamaIndex analysis: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log("LlamaIndex analysis completed successfully.")

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "LlamaIndex analysis completed successfully",
		"scanData":         scanData,
		"analysisResponse": llamaResponse,
		"query":            query,
	})
}

// New implementation to replace remediateWithOllamaHandler
func remediateHandler(w http.ResponseWriter, r *http.Request) {
	sbomFile := appConfig.SBOMOutputFile

	if _, err := os.Stat(sbomFile); os.IsNotExist(err) {
		http.Error(w, "SBOM file not found. Please generate it first.", http.StatusBadRequest)
		return
	}

	logger.Log("Starting remediation...")

	// Run Grype scan to get output
	scanOutput, err := runGrypeScan(sbomFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Error running Grype for remediation: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(scanOutput) == 0 {
		w.Header().Set("Content-Type", contentTypeJSON)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":           "No vulnerabilities found that need remediation",
			"remediationScript": "",
		})
		return
	}

	// Try LlamaIndex first
	client := NewLlamaIndexClient(appConfig.LlamaIndexEndpoint)
	sbomContent, err := os.ReadFile(sbomFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to read SBOM file: %v", err))
		http.Error(w, "Failed to read SBOM file", http.StatusInternalServerError)
		return
	}
	llamaResponse, err := client.AnalyzeVulnerabilities(scanOutput, string(sbomContent))

	if err == nil {
		logger.Log("Remediation script generated using LlamaIndex.")
		w.Header().Set("Content-Type", contentTypeJSON)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":           "Remediation script generated successfully using LlamaIndex",
			"remediationScript": llamaResponse,
			"engine":            "llamaindex",
		})
		return
	}

	logger.Log(fmt.Sprintf("LlamaIndex analysis failed: %v. Falling back to Ollama.", err))

	// Fallback to Ollama
	pkgType := detectPackageType(scanOutput)
	ollamaResponse, err := getOllamaRemediation(scanOutput, pkgType)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get Ollama remediation: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log("Remediation script generated using Ollama fallback.")

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Remediation script generated successfully using Ollama",
		"remediationScript": ollamaResponse,
		"engine":            "ollama",
		"ollamaModel":       appConfig.DefaultModel,
	})
}

// Check if Ollama is available
func checkOllamaAvailability() error {
	// Ollama uses /api/generate for its API endpoint
	url := appConfig.OllamaHost + "/api/generate"

	payload := map[string]interface{}{
		"model":  appConfig.DefaultModel,
		"prompt": "test",
		"stream": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create test payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama service error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Get remediation script from Ollama
func getOllamaRemediation(scanOutput string, pkgType string) (string, error) {
	// Check Ollama availability first
	if err := checkOllamaAvailability(); err != nil {
		logger.Log(fmt.Sprintf("ollama availability check failed: %v", err))
		return "", fmt.Errorf("ollama service is not available, please ensure ollama is running and the model '%s' is installed: %w", appConfig.DefaultModel, err)
	}

	prompt := fmt.Sprintf(`You are a DevSecOps expert. Given the following SBOM scan output, write a clean script that upgrades each vulnerable %s to its fixed version.

Only output the script in a code block.

SBOM Scan:
%s
`, pkgType, scanOutput)

	logger.Log(fmt.Sprintf("Using Ollama model: %s", appConfig.DefaultModel))

	llm, err := ollama.New(
		ollama.WithModel(appConfig.DefaultModel),
		ollama.WithServerURL(appConfig.OllamaHost),
	)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Ollama client: %w", err)
	}

	response, err := llm.Call(context.Background(), prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get response from Ollama LLM: %w", err)
	}

	return response, nil
}

// Detect package type from scan output
func detectPackageType(scanOutput string) string {
	pkgType := "package" // default fallback

	if strings.Contains(scanOutput, "python") {
		pkgType = "Python package"
	} else if strings.Contains(scanOutput, "nodejs") || strings.Contains(scanOutput, "npm") {
		pkgType = "Node.js package"
	} else if strings.Contains(scanOutput, "java") || strings.Contains(scanOutput, "maven") {
		pkgType = "Java package"
	} else if strings.Contains(scanOutput, "golang") || strings.Contains(scanOutput, "go-module") {
		pkgType = "Go package"
	} else if strings.Contains(scanOutput, "ruby") || strings.Contains(scanOutput, "gem") {
		pkgType = "Ruby gem"
	} else if strings.Contains(scanOutput, "rust") || strings.Contains(scanOutput, "cargo") {
		pkgType = "Rust crate"
	}

	return pkgType
}

// Clone a Git repository
func cloneGitRepo(repoURL string, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove existing git clone directory: %w", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed in the container: %w", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone git repository: %w", err)
	}
	return nil
}

// Get all source tags
func allSourceTags() []string {
	return collections.TaggedValueSet[source.Provider]{}.Join(sourceproviders.All("", nil)...).Tags()
}

// Save SBOM to file
func saveSBOMToFile(s *sbom.SBOM, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create SBOM file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Log(fmt.Sprintf("Error closing SBOM file: %v", closeErr))
		}
	}()

	encoder, err := cyclonedxjson.NewFormatEncoderWithConfig(cyclonedxjson.DefaultEncoderConfig())
	if err != nil {
		return fmt.Errorf("failed to create CycloneDX encoder: %w", err)
	}

	if err := encoder.Encode(file, *s); err != nil {
		return fmt.Errorf("failed to encode SBOM: %w", err)
	}
	return nil
}
