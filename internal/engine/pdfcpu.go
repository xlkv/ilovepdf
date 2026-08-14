package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type PDFCPUEngine struct{}

func NewPDFCPUEngine() *PDFCPUEngine {
	return &PDFCPUEngine{}
}

// MergePDFs merges multiple input PDF files into outputFilePath
func (e *PDFCPUEngine) MergePDFs(inputPaths []string, outputFilePath string) error {
	conf := model.NewDefaultConfiguration()
	return api.MergeCreateFile(inputPaths, outputFilePath, false, conf)
}

// SplitPDF extracts specific pages or splits all pages into outDir
func (e *PDFCPUEngine) SplitPDF(inputPath, outDir string, pages []string) ([]string, error) {
	conf := model.NewDefaultConfiguration()
	if len(pages) > 0 {
		pageSelection := strings.Join(pages, ",")
		err := api.ExtractPagesFile(inputPath, outDir, []string{pageSelection}, conf)
		if err != nil {
			return nil, err
		}
	} else {
		err := api.SplitFile(inputPath, outDir, 1, conf)
		if err != nil {
			return nil, err
		}
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		return nil, err
	}

	var outputFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			outputFiles = append(outputFiles, filepath.Join(outDir, f.Name()))
		}
	}
	return outputFiles, nil
}

// RotatePDF rotates pages by degree (90, 180, 270)
func (e *PDFCPUEngine) RotatePDF(inputPath, outputPath string, degrees int) error {
	conf := model.NewDefaultConfiguration()
	return api.RotateFile(inputPath, outputPath, degrees, nil, conf)
}

// EncryptPDF protects PDF with user password
func (e *PDFCPUEngine) EncryptPDF(inputPath, outputPath, userPw string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPw
	conf.OwnerPW = userPw + "_owner"
	return api.EncryptFile(inputPath, outputPath, conf)
}

// DecryptPDF unlocks PDF using user password
func (e *PDFCPUEngine) DecryptPDF(inputPath, outputPath, userPw string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPw
	return api.DecryptFile(inputPath, outputPath, conf)
}

// WatermarkPDF adds a text watermark to PDF
func (e *PDFCPUEngine) WatermarkPDF(inputPath, outputPath, text string) error {
	conf := model.NewDefaultConfiguration()
	wm, err := api.TextWatermark(text, "font:Helvetica, scale:0.7, points:48, op:0.4", true, false, types.POINTS)
	if err != nil {
		return err
	}
	return api.AddWatermarksFile(inputPath, outputPath, nil, wm, conf)
}

// AddPageNumbers adds page numbers to bottom right
func (e *PDFCPUEngine) AddPageNumbers(inputPath, outputPath string) error {
	conf := model.NewDefaultConfiguration()
	np, err := api.TextWatermark("%p of %P", "pos:br, font:Helvetica, points:12, op:1.0", false, false, types.POINTS)
	if err != nil {
		return err
	}
	return api.AddWatermarksFile(inputPath, outputPath, nil, np, conf)
}

// OrganizePDF reorders or deletes pages
func (e *PDFCPUEngine) OrganizePDF(inputPath, outputPath string, selectedPages []string) error {
	conf := model.NewDefaultConfiguration()
	return api.CollectFile(inputPath, outputPath, selectedPages, conf)
}

// ImagesToPDF converts image files to PDF
func (e *PDFCPUEngine) ImagesToPDF(imagePaths []string, outputPath string) error {
	conf := model.NewDefaultConfiguration()
	imp, err := api.Import("pos:c, sc:1.0", types.POINTS)
	if err != nil {
		return err
	}
	return api.ImportImagesFile(imagePaths, outputPath, imp, conf)
}

// GetPageCount returns total page count of PDF
func (e *PDFCPUEngine) GetPageCount(inputPath string) (int, error) {
	return api.PageCountFile(inputPath)
}

// Helper parsing range string e.g. "1-3, 5, 8"
func ParsePageRange(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func FormatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024.0)
	}
	return fmt.Sprintf("%.2f MB", float64(size)/(1024.0*1024.0))
}

func ParseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
