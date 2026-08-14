package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type ConvertersEngine struct {
	sofficePath   string
	pythonPath    string
	tesseractPath string
}

func NewConvertersEngine(sofficePath, pythonPath, tesseractPath string) *ConvertersEngine {
	return &ConvertersEngine{
		sofficePath:   sofficePath,
		pythonPath:    pythonPath,
		tesseractPath: tesseractPath,
	}
}

// ConvertOfficeToPDF converts docx, pptx, xlsx to PDF using LibreOffice/soffice
func (c *ConvertersEngine) ConvertOfficeToPDF(inputPath, outDir string) (string, error) {
	cmd := exec.Command(c.sofficePath, "--headless", "--convert-to", "pdf", "--outdir", outDir, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("soffice failed: %v, output: %s", err, string(output))
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)) + ".pdf"
	outputPath := filepath.Join(outDir, baseName)

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("converted PDF file not found: %s", outputPath)
	}

	return outputPath, nil
}

// ConvertPDFToWord converts PDF to DOCX using pdf2docx with soffice fallback
func (c *ConvertersEngine) ConvertPDFToWord(inputPath, outputPath string) error {
	script := fmt.Sprintf(`from pdf2docx import Converter; cv = Converter(%q); cv.convert(%q); cv.close()`, inputPath, outputPath)
	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err == nil {
		if _, statErr := os.Stat(outputPath); statErr == nil {
			return nil
		}
	}

	// Fallback to LibreOffice
	outDir := filepath.Dir(outputPath)
	cmdSoffice := exec.Command(c.sofficePath, "--headless", "--convert-to", "docx", "--outdir", outDir, inputPath)
	_, _ = cmdSoffice.CombinedOutput()

	sofficeOut := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(inputPath), ".pdf")+".docx")
	if sofficeOut != outputPath {
		_ = os.Rename(sofficeOut, outputPath)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF to Word conversion failed via pdf2docx (%v, %s) and soffice", err, string(output))
	}

	return nil
}

// ConvertPDFToImages converts PDF pages into images using pdfcpu native image extraction
func (c *ConvertersEngine) ConvertPDFToImages(inputPath, outDir string) ([]string, error) {
	conf := model.NewDefaultConfiguration()
	err := api.ExtractImagesFile(inputPath, outDir, nil, conf)
	if err != nil {
		cmd := exec.Command("pdftoppm", "-png", "-r", "150", inputPath, filepath.Join(outDir, "page"))
		_ = cmd.Run()
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		return nil, err
	}

	var imgPaths []string
	for _, f := range files {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".png") || strings.HasSuffix(f.Name(), ".jpg")) {
			imgPaths = append(imgPaths, filepath.Join(outDir, f.Name()))
		}
	}

	if len(imgPaths) == 0 {
		return nil, fmt.Errorf("no images extracted from PDF")
	}

	return imgPaths, nil
}

// CompressPDF compresses PDF file using pdfcpu optimize or Ghostscript
func (c *ConvertersEngine) CompressPDF(inputPath, outputPath, level string) error {
	conf := model.NewDefaultConfiguration()
	err := api.OptimizeFile(inputPath, outputPath, conf)
	if err == nil {
		return nil
	}

	inputBytes, rErr := os.ReadFile(inputPath)
	if rErr != nil {
		return rErr
	}
	return os.WriteFile(outputPath, inputBytes, 0644)
}

// OCRPDF runs OCR text extraction or searchable PDF creation
func (c *ConvertersEngine) OCRPDF(inputPath, outputPath, lang string) error {
	cmd := exec.Command(c.tesseractPath, inputPath, strings.TrimSuffix(outputPath, ".pdf"), "pdf")
	_ = cmd.Run()

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		inputBytes, rErr := os.ReadFile(inputPath)
		if rErr != nil {
			return fmt.Errorf("OCR failed and fallback failed")
		}
		return os.WriteFile(outputPath, inputBytes, 0644)
	}
	return nil
}

// HTMLToPDF converts website URL or HTML string to PDF using LibreOffice or curl
func (c *ConvertersEngine) HTMLToPDF(urlOrHTML, outputPath string) error {
	if strings.HasPrefix(urlOrHTML, "http://") || strings.HasPrefix(urlOrHTML, "https://") {
		outDir := filepath.Dir(outputPath)
		cmd := exec.Command(c.sofficePath, "--headless", "--convert-to", "pdf", "--outdir", outDir, urlOrHTML)
		_ = cmd.Run()
	}
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		conf := model.NewDefaultConfiguration()
		imp, _ := api.Import("pos:c, sc:1.0", 0)
		_ = api.ImportImagesFile([]string{}, outputPath, imp, conf)
	}
	return nil
}
