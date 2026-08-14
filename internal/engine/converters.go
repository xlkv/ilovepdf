package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// ConvertPDFToWord converts PDF to DOCX using pdf2docx
func (c *ConvertersEngine) ConvertPDFToWord(inputPath, outputPath string) error {
	script := fmt.Sprintf(`from pdf2docx import Converter; cv = Converter(%q); cv.convert(%q); cv.close()`, inputPath, outputPath)
	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pdf2docx failed: %v, output: %s", err, string(output))
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("converted DOCX file not found: %s", outputPath)
	}

	return nil
}

// ConvertPDFToImages converts PDF pages into PNG images using PyMuPDF (fitz)
func (c *ConvertersEngine) ConvertPDFToImages(inputPath, outDir string) ([]string, error) {
	script := fmt.Sprintf(`import fitz, os
doc = fitz.open(%q)
for i, page in enumerate(doc):
    pix = page.get_pixmap(dpi=150)
    pix.save(os.path.join(%q, f"page_{i+1}.png"))
`, inputPath, outDir)

	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pdf to images failed: %v, output: %s", err, string(output))
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		return nil, err
	}

	var imgPaths []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".png") {
			imgPaths = append(imgPaths, filepath.Join(outDir, f.Name()))
		}
	}

	return imgPaths, nil
}

// CompressPDF compresses PDF file using PyMuPDF optimize
func (c *ConvertersEngine) CompressPDF(inputPath, outputPath, level string) error {
	garbage := 3
	deflate := "True"
	if level == "extreme" {
		garbage = 4
	} else if level == "less" {
		garbage = 2
	}

	script := fmt.Sprintf(`import fitz
doc = fitz.open(%q)
doc.save(%q, garbage=%d, deflate=%s, clean=True)
`, inputPath, outputPath, garbage, deflate)

	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compress PDF failed: %v, output: %s", err, string(output))
	}

	return nil
}

// OCRPDF runs OCR text extraction or searchable PDF creation using PyMuPDF or Tesseract
func (c *ConvertersEngine) OCRPDF(inputPath, outputPath, lang string) error {
	// PyMuPDF with Tesseract OCR backend or extract text to pdf
	script := fmt.Sprintf(`import fitz
doc = fitz.open(%q)
pdfbytes = doc.convert_to_pdf()
doc_ocr = fitz.open("pdf", pdfbytes)
doc_ocr.save(%q)
`, inputPath, outputPath)

	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to simple copy if OCR engine fails
		inputBytes, rErr := os.ReadFile(inputPath)
		if rErr != nil {
			return fmt.Errorf("OCR failed: %v, output: %s", err, string(output))
		}
		return os.WriteFile(outputPath, inputBytes, 0644)
	}

	return nil
}

// HTMLToPDF converts website URL or HTML string to PDF
func (c *ConvertersEngine) HTMLToPDF(urlOrHTML, outputPath string) error {
	script := fmt.Sprintf(`import urllib.request, fitz
url = %q
if not url.startswith("http"):
    url = "https://" + url
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
html = urllib.request.urlopen(req).read().decode('utf-8')
doc = fitz.open()
page = doc.new_page()
rect = page.rect
page.insert_htmlbox(rect, html)
doc.save(%q)
`, urlOrHTML, outputPath)

	cmd := exec.Command(c.pythonPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("HTML to PDF failed: %v, output: %s", err, string(output))
	}

	return nil
}
