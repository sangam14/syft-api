package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"context"

	"github.com/anchore/go-collections"
	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
	"github.com/anchore/syft/syft/format/cyclonedxjson"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/anchore/syft/syft/source/sourceproviders"
	"github.com/gofiber/fiber/v2"
	"github.com/tmc/langchaingo/llms/ollama"
)

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

func writeToLogFile(message string) {
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("%s %s", currentTime, message)
	file, err := os.OpenFile("static/output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer file.Close()
	logger := log.New(file, "", log.LstdFlags)
	logger.Println(logMessage)
}

func main() {

	app := fiber.New()

	app.Static("/", "./static") // serving static UI files from the static/ directory
	fmt.Println("Serving static files from ./static")
	app.Get("/generate-sbom", generateSBOM)
	app.Get("/scan-sbom", scanSBOM)
	app.Get("/logs", logsHandler)
	app.Get("/remediate", remediateWithOllama)

	fmt.Println("API is running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}

func generateSBOM(c *fiber.Ctx) error {
	source := c.Query("source")

	if source == "" {
		msg := "Error: No valid source provided. Provide an image, directory path, or remote URL."
		writeToLogFile(msg)
		return c.Status(400).JSON(fiber.Map{"error": msg})
	}

	var sourceInput string
	if _, err := os.Stat(source); err == nil {
		sourceInput = "dir:" + source
	} else if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		cloneDir := "/tmp/git-sbom"
		if err := cloneGitRepo(source, cloneDir); err != nil {
			msg := fmt.Sprintf("Failed to clone repository: %v", err)
			writeToLogFile(msg)
			return c.Status(500).JSON(fiber.Map{"error": msg})
		}
		sourceInput = "dir:" + cloneDir
	} else {
		sourceInput = "image:" + source
	}

	writeToLogFile(fmt.Sprintf("Processing SBOM for source: %s", sourceInput))

	schemeSource, newUserInput := stereoscope.ExtractSchemeSource(sourceInput, allSourceTags()...)
	getSourceCfg := syft.DefaultGetSourceConfig()
	if schemeSource != "" {
		getSourceCfg = getSourceCfg.WithSources(schemeSource)
		sourceInput = newUserInput
	}

	src, err := syft.GetSource(context.Background(), sourceInput, getSourceCfg)
	if err != nil {
		msg := fmt.Sprintf("Failed to get source: %v", err)
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
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
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
	}

	// Save SBOM to file
	sbomFile := "sbom.cyclonedx.json"
	saveSBOMToFile(sbomData, sbomFile)

	// Read SBOM content for response
	sbomContent, err := os.ReadFile(sbomFile)
	if err != nil {
		msg := "Failed to read generated SBOM file."
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
	}

	writeToLogFile("SBOM generated successfully.")

	return c.JSON(fiber.Map{
		"message":  "SBOM generated successfully",
		"format":   "CycloneDX JSON",
		"file":     sbomFile,
		"sbomData": string(sbomContent),
	})
}

func scanSBOM(c *fiber.Ctx) error {
	sbomFile := "sbom.cyclonedx.json"

	if _, err := os.Stat(sbomFile); os.IsNotExist(err) {
		return c.Status(400).JSON(fiber.Map{"error": "SBOM file not found. Please generate it first."})
	}

	writeToLogFile("Starting SBOM scan...")

	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running Grype: %v", err)
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
	}

	scanOutput := string(output)
	pkgType := "package" // default fallback
	writeToLogFile(fmt.Sprintf("Detected package type: %s", pkgType))
	if strings.Contains(scanOutput, "python") {
		pkgType = "Python package"
	} else if strings.Contains(scanOutput, "nodejs") || strings.Contains(scanOutput, "npm") {
		pkgType = "Node.js package"
	} else if strings.Contains(scanOutput, "java") || strings.Contains(scanOutput, "maven") {
		pkgType = "Java package"
	}

	prompt := fmt.Sprintf(`You are a DevSecOps expert. Given the following SBOM scan output, write a clean script that upgrades each vulnerable %s to its fixed version.

Only output the script in a code block.

SBOM Scan:
%s
`, pkgType, scanOutput)

	writeToLogFile(fmt.Sprintf("Using Ollama model: %s", "mistral"))

	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		writeToLogFile(fmt.Sprintf("Failed to initialize Ollama client: %v", err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	response, err := llm.Call(context.Background(), prompt)
	if err != nil {
		writeToLogFile(fmt.Sprintf("Failed to get response from Ollama LLM: %v", err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	llmResponse := response

	writeToLogFile("SBOM scan completed successfully.")

	return c.JSON(fiber.Map{
		"message":             "Grype scan and remediation completed successfully",
		"scanResult":          string(output),
		"remediationScript":   llmResponse,
		"remediationCommands": extractScriptBlock(llmResponse),
		"pkgType":             pkgType,
		"markdownResponse":    fmt.Sprintf("```bash\n%s\n```", extractScriptBlock(llmResponse)),
		"ollamaModel":         "mistral",
		"ollamaRawResponse":   llmResponse,
	})
}

func logsHandler(c *fiber.Ctx) error {
	content, err := os.ReadFile("static/output.log")
	if err != nil {
		return c.Status(500).SendString("Failed to read logs")
	}
	return c.SendString(string(content))
}

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

func allSourceTags() []string {
	return collections.TaggedValueSet[source.Provider]{}.Join(sourceproviders.All("", nil)...).Tags()
}

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

func remediateWithOllama(c *fiber.Ctx) error {
	sbomFile := "sbom.cyclonedx.json"

	if _, err := os.Stat(sbomFile); os.IsNotExist(err) {
		return c.Status(400).JSON(fiber.Map{"error": "SBOM file not found. Please generate it first."})
	}

	writeToLogFile("Starting remediation using Ollama...")

	// Run Grype scan to get output
	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running Grype for remediation: %v", err)
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
	}

	modelName := "mistral"
	payload := map[string]string{
		"model":  modelName,
		"prompt": fmt.Sprintf("Act like a security expert and write a shell script to fix these vulnerabilities:\n\n%s", string(output)),
	}
	payloadBytes, _ := json.Marshal(payload)
	writeToLogFile(fmt.Sprintf("Using Ollama model: %s", modelName))

	// Call Ollama API
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://host.docker.internal:11434"
	}
	url := fmt.Sprintf("%s/api/generate", ollamaHost)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		writeToLogFile(fmt.Sprintf("Failed to create request: %v", err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeToLogFile(fmt.Sprintf("Failed to get response from Ollama API: %v", err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	var remediationBuilder strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		writeToLogFile("Ollama Stream Line: " + line)
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			writeToLogFile(fmt.Sprintf("Failed to parse Ollama response line: %v", err))
			continue
		}
		if msg, ok := result["message"].(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				remediationBuilder.WriteString(content)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		writeToLogFile(fmt.Sprintf("Scanner error reading Ollama response: %v", err))
		return c.Status(500).JSON(fiber.Map{"error": "Error reading Ollama response stream"})
	}
	if remediationBuilder.Len() == 0 {
		writeToLogFile("No response content from Ollama. Setting default remediation message.")
		remediationBuilder.WriteString("No remediation script generated. Please verify the prompt or model.")
	}
	remediationText := remediationBuilder.String()

	writeToLogFile("Remediation script generated using Ollama.")

	return c.JSON(fiber.Map{
		"message":           "Remediation script generated successfully",
		"remediationScript": remediationText,
		"ollamaModel":       modelName,
		"ollamaRawResponse": remediationText,
	})
}
