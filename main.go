package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/gofiber/fiber/v2"
)

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

	writeToLogFile("Starting SBOM scan...")

	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running Grype: %v", err)
		writeToLogFile(msg)
		return c.Status(500).JSON(fiber.Map{"error": msg})
	}

	writeToLogFile("SBOM scan completed successfully.")

	return c.JSON(fiber.Map{
		"message":    "Grype scan completed successfully",
		"scanResult": string(output),
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
