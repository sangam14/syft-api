package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	LlamaIndexEndpoint: getEnv("LLAMA_INDEX_ENDPOINT", "http://llama-index-api:8000"),
	OllamaHost:         getEnv("OLLAMA_HOST", "http://host.docker.internal:11434"),
	DefaultModel:       getEnv("DEFAULT_MODEL", "mistral"),
	LogFile:            "static/output.log",
	SBOMOutputFile:     "sbom.cyclonedx.json",
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
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(c.BaseURL+"/query", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to call LlamaIndex API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LlamaIndex API returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse LlamaIndex response: %v", err)
	}

	if response, ok := result["response"].(string); ok {
		return response, nil
	}

	return "", fmt.Errorf("invalid response format from LlamaIndex")
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
		return nil, fmt.Errorf("failed to open log file: %v", err)
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
		l.file.Close()
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

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static"))).Methods("GET")
	fmt.Println("Serving static files from ./static")

	// API routes
	r.HandleFunc("/generate-sbom", generateSBOMHandler).Methods("POST")
	r.HandleFunc("/scan-sbom", scanSBOMHandler).Methods("POST")
	r.HandleFunc("/logs", logsHandler).Methods("GET")
	r.HandleFunc("/remediate", remediateHandler).Methods("GET")
	r.HandleFunc("/llamaindex-analyze", llamaIndexAnalyzeHandler).Methods("POST")
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	port := getEnv("PORT", "3000")
	fmt.Printf("API is running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// Health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var sourceInput string
	if _, err := os.Stat(source); err == nil {
		sourceInput = "dir:" + source
	} else if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		cloneDir := "/tmp/git-sbom"
		if err := cloneGitRepo(source, cloneDir); err != nil {
			msg := fmt.Sprintf("Failed to clone repository: %v", err)
			logger.Log(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		sourceInput = "dir:" + cloneDir
	} else {
		sourceInput = "image:" + source
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
	saveSBOMToFile(sbomData, appConfig.SBOMOutputFile)

	// Read SBOM content for response
	sbomContent, err := os.ReadFile(appConfig.SBOMOutputFile)
	if err != nil {
		msg := "Failed to read generated SBOM file."
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	logger.Log("SBOM generated successfully.")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "SBOM generated successfully",
		"format":   "CycloneDX JSON",
		"file":     appConfig.SBOMOutputFile,
		"sbomData": string(sbomContent),
	})
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

	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running Grype: %v", err)
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	scanOutput := string(output)
	pkgType := detectPackageType(scanOutput)
	logger.Log(fmt.Sprintf("Detected package type: %s", pkgType))

	// Read SBOM content for enhanced analysis
	sbomContent, err := os.ReadFile(sbomFile)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to read SBOM file: %v", err))
	}

	remediation := ""
	if body.UseAdvanced && len(scanOutput) > 0 {
		// Use LlamaIndex for enhanced analysis
		client := NewLlamaIndexClient(appConfig.LlamaIndexEndpoint)
		llamaResponse, err := client.AnalyzeVulnerabilities(scanOutput, string(sbomContent))
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get LlamaIndex analysis: %v, falling back to Ollama", err))
			remediation, err = getOllamaRemediation(scanOutput, pkgType)
			if err != nil {
				logger.Log(fmt.Sprintf("Failed to get Ollama remediation: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			remediation = llamaResponse
			logger.Log("Used LlamaIndex for enhanced vulnerability analysis.")
		}
	} else if len(scanOutput) > 0 {
		// Use Ollama for standard remediation
		remediation, err = getOllamaRemediation(scanOutput, pkgType)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get Ollama remediation: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		remediation = "No vulnerabilities found that need remediation."
	}

	logger.Log("SBOM scan completed successfully.")

	w.Header().Set("Content-Type", "application/json")
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
	})
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(appConfig.LogFile)
	if err != nil {
		http.Error(w, "Failed to read logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
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
		cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
		output, err := cmd.CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("Error running Grype: %v", err)
			logger.Log(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		scanData = string(output)
	}

	logger.Log("Running LlamaIndex analysis...")
	llamaResponse, err := client.AnalyzeVulnerabilities(scanData, string(sbomContent))
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get LlamaIndex analysis: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log("LlamaIndex analysis completed successfully.")

	w.Header().Set("Content-Type", "application/json")
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
	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running Grype for remediation: %v", err)
		logger.Log(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	scanOutput := string(output)
	if len(scanOutput) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":           "No vulnerabilities found that need remediation",
			"remediationScript": "",
		})
		return
	}

	// Try LlamaIndex first
	client := NewLlamaIndexClient(appConfig.LlamaIndexEndpoint)
	sbomContent, _ := os.ReadFile(sbomFile)
	llamaResponse, err := client.AnalyzeVulnerabilities(scanOutput, string(sbomContent))

	if err == nil {
		logger.Log("Remediation script generated using LlamaIndex.")
		w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Remediation script generated successfully using Ollama",
		"remediationScript": ollamaResponse,
		"engine":            "ollama",
		"ollamaModel":       appConfig.DefaultModel,
	})
}

// Get remediation script from Ollama
func getOllamaRemediation(scanOutput string, pkgType string) (string, error) {
	prompt := fmt.Sprintf(`You are a DevSecOps expert. Given the following SBOM scan output, write a clean script that upgrades each vulnerable %s to its fixed version.

Only output the script in a code block.

SBOM Scan:
%s
`, pkgType, scanOutput)

	logger.Log(fmt.Sprintf("Using Ollama model: %s", appConfig.DefaultModel))

	llm, err := ollama.New(ollama.WithModel(appConfig.DefaultModel), ollama.WithServerURL(appConfig.OllamaHost))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Ollama client: %v", err)
	}

	response, err := llm.Call(context.Background(), prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get response from Ollama LLM: %v", err)
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
	os.RemoveAll(dest)

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed in the container: %v", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Get all source tags
func allSourceTags() []string {
	return collections.TaggedValueSet[source.Provider]{}.Join(sourceproviders.All("", nil)...).Tags()
}

// Save SBOM to file
func saveSBOMToFile(s *sbom.SBOM, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	encoder, err := cyclonedxjson.NewFormatEncoderWithConfig(cyclonedxjson.DefaultEncoderConfig())
	if err != nil {
		return
	}

	encoder.Encode(file, *s)
}
