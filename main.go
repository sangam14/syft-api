package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/anchore/go-collections"
	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
	"github.com/anchore/syft/syft/format/cyclonedxjson" // ✅ Correct import
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/anchore/syft/syft/source/sourceproviders"
	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // Swagger UI
)

// @title SBOM & Vulnerability Scanner API
// @version 1.0
// @description API for generating SBOMs in CycloneDX format and scanning for vulnerabilities using Syft & Grype.
// @host localhost:3000
// @BasePath /
// @schemes http
func main() {
	app := fiber.New()

	// Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault) // http://localhost:3000/swagger/index.html

	// API routes
	app.Get("/generate-sbom/:image", generateSBOM)
	app.Get("/scan-sbom", scanSBOM)

	// Start server
	fmt.Println("🚀 API is running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}

// @Summary Generate an SBOM in CycloneDX JSON format
// @Description Creates an SBOM for the given Docker image and saves it in CycloneDX JSON format
// @Produce json
// @Param image path string true "Docker Image Name"
// @Success 200 {string} string "SBOM generated successfully"
// @Failure 500 {string} string "Failed to generate SBOM"
// @Router /generate-sbom/{image} [get]
func generateSBOM(c *fiber.Ctx) error {
	image := c.Params("image", "vulnerables/web-dvwa")

	schemeSource, newUserInput := stereoscope.ExtractSchemeSource(image, allSourceTags()...)
	getSourceCfg := syft.DefaultGetSourceConfig()
	if schemeSource != "" {
		getSourceCfg = getSourceCfg.WithSources(schemeSource)
		image = newUserInput
	}

	src, err := syft.GetSource(context.Background(), image, getSourceCfg)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to get source: %v", err))
	}

	cfg := syft.DefaultCreateSBOMConfig().WithCatalogerSelection(
		pkgcataloging.NewSelectionRequest().WithDefaults(pkgcataloging.InstalledTag, pkgcataloging.DirectoryTag, pkgcataloging.ImageTag),
	)
	sbomData, err := syft.CreateSBOM(context.Background(), src, cfg)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to create SBOM: %v", err))
	}

	saveSBOMToFile(sbomData, "sbom.cyclonedx.json")

	return c.JSON(fiber.Map{
		"message": "✅ SBOM generated successfully",
		"format":  "CycloneDX JSON",
		"file":    "sbom.cyclonedx.json",
	})
}

// @Summary Scan the SBOM for vulnerabilities
// @Description Runs Grype to scan the generated SBOM for vulnerabilities (quiet mode)
// @Produce json
// @Success 200 {string} string "Scan completed successfully"
// @Failure 500 {string} string "Error running Grype"
// @Router /scan-sbom [get]
func scanSBOM(c *fiber.Ctx) error {
	sbomFile := "sbom.cyclonedx.json"
	fmt.Println("\n🔍 Running vulnerability scan using Grype (Quiet Mode)...")

	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q") // ✅ Run Grype in quiet mode
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("❌ Error running Grype: %v", err))
	}

	return c.JSON(fiber.Map{
		"message": "✅ Grype scan completed successfully!",
	})
}

func allSourceTags() []string {
	return collections.TaggedValueSet[source.Provider]{}.Join(sourceproviders.All("", nil)...).Tags()
}

// ✅ Corrected SBOM save function
func saveSBOMToFile(s *sbom.SBOM, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create SBOM file: %v\n", err)
		return
	}
	defer file.Close()

	// ✅ Get the CycloneDX JSON format encoder
	encoder, err := cyclonedxjson.NewFormatEncoderWithConfig(cyclonedxjson.DefaultEncoderConfig())
	if err != nil {
		fmt.Printf("Failed to create CycloneDX encoder: %v\n", err)
		return
	}

	// ✅ Encode SBOM to CycloneDX JSON format
	if err := encoder.Encode(file, *s); err != nil {
		fmt.Printf("Failed to encode SBOM: %v\n", err)
	}
}
