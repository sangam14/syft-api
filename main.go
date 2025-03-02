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
	"github.com/anchore/syft/syft/format/cyclonedxjson"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/anchore/syft/syft/source/sourceproviders"
	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

func main() {
	app := fiber.New()

	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Get("/generate-sbom", generateSBOM)
	app.Get("/scan-sbom", scanSBOM)

	fmt.Println("API is running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}

func generateSBOM(c *fiber.Ctx) error {
	image := c.Query("image")
	dir := c.Query("dir")
	remote := c.Query("remote")

	var sourceInput string
	if image != "" {
		sourceInput = "image:" + image
	} else if dir != "" {
		sourceInput = "dir:" + dir
	} else if remote != "" {
		cloneDir := "/tmp/git-sbom"
		if err := cloneGitRepo(remote, cloneDir); err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to clone repository: %v", err))
		}
		sourceInput = "dir:" + cloneDir
	} else {
		return c.Status(400).SendString("Error: No valid source provided. Use 'image', 'dir', or 'remote'.")
	}

	schemeSource, newUserInput := stereoscope.ExtractSchemeSource(sourceInput, allSourceTags()...)
	getSourceCfg := syft.DefaultGetSourceConfig()
	if schemeSource != "" {
		getSourceCfg = getSourceCfg.WithSources(schemeSource)
		sourceInput = newUserInput
	}

	src, err := syft.GetSource(context.Background(), sourceInput, getSourceCfg)
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
		"message": "SBOM generated successfully",
		"format":  "CycloneDX JSON",
		"file":    "sbom.cyclonedx.json",
	})
}

func scanSBOM(c *fiber.Ctx) error {
	sbomFile := "sbom.cyclonedx.json"

	cmd := exec.Command("grype", sbomFile, "--only-fixed", "-q")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Error running Grype: %v", err))
	}

	return c.JSON(fiber.Map{
		"message": "Grype scan completed successfully",
	})
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
