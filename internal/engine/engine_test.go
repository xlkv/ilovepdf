package engine

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createSamplePNG(filePath string, c color.Color) error {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestPDFEngineAllCases(t *testing.T) {
	tempDir := t.TempDir()
	pdfcpu := NewPDFCPUEngine()
	conv := NewConvertersEngine("/usr/local/bin/soffice", "/usr/local/bin/python3", "/opt/homebrew/bin/tesseract")

	// 1. Create 2 sample images natively in Go and convert to PDF
	img1 := filepath.Join(tempDir, "img1.png")
	img2 := filepath.Join(tempDir, "img2.png")

	if err := createSamplePNG(img1, color.RGBA{R: 255, A: 255}); err != nil {
		t.Fatalf("Failed to create img1: %v", err)
	}
	if err := createSamplePNG(img2, color.RGBA{B: 255, A: 255}); err != nil {
		t.Fatalf("Failed to create img2: %v", err)
	}

	pdf1 := filepath.Join(tempDir, "doc1.pdf")
	err := pdfcpu.ImagesToPDF([]string{img1, img2}, pdf1)
	if err != nil {
		t.Fatalf("ImagesToPDF failed: %v", err)
	}

	// 2. Test GetPageCount
	count, err := pdfcpu.GetPageCount(pdf1)
	if err != nil || count != 2 {
		t.Fatalf("GetPageCount failed: count=%d, err=%v", count, err)
	}

	// 3. Test Watermark
	pdfWatermarked := filepath.Join(tempDir, "watermarked.pdf")
	err = pdfcpu.WatermarkPDF(pdf1, pdfWatermarked, "TEST WATERMARK")
	if err != nil {
		t.Fatalf("WatermarkPDF failed: %v", err)
	}

	// 4. Test AddPageNumbers
	pdfNumbered := filepath.Join(tempDir, "numbered.pdf")
	err = pdfcpu.AddPageNumbers(pdf1, pdfNumbered)
	if err != nil {
		t.Fatalf("AddPageNumbers failed: %v", err)
	}

	// 5. Test Rotate
	pdfRotated := filepath.Join(tempDir, "rotated.pdf")
	err = pdfcpu.RotatePDF(pdf1, pdfRotated, 90)
	if err != nil {
		t.Fatalf("RotatePDF failed: %v", err)
	}

	// 6. Test Protect & Unlock
	pdfProtected := filepath.Join(tempDir, "protected.pdf")
	err = pdfcpu.EncryptPDF(pdf1, pdfProtected, "secret123")
	if err != nil {
		t.Fatalf("EncryptPDF failed: %v", err)
	}

	pdfUnlocked := filepath.Join(tempDir, "unlocked.pdf")
	err = pdfcpu.DecryptPDF(pdfProtected, pdfUnlocked, "secret123")
	if err != nil {
		t.Fatalf("DecryptPDF failed: %v", err)
	}

	// 7. Test Merge
	pdfMerged := filepath.Join(tempDir, "merged.pdf")
	err = pdfcpu.MergePDFs([]string{pdf1, pdfNumbered}, pdfMerged)
	if err != nil {
		t.Fatalf("MergePDFs failed: %v", err)
	}

	// 8. Test Split
	splitDir := filepath.Join(tempDir, "split_out")
	_ = os.MkdirAll(splitDir, 0755)
	splitFiles, err := pdfcpu.SplitPDF(pdfMerged, splitDir, []string{"1"})
	if err != nil || len(splitFiles) == 0 {
		t.Fatalf("SplitPDF failed: %v", err)
	}

	// 9. Test PDF to Images
	imgOutDir := filepath.Join(tempDir, "imgs_out")
	_ = os.MkdirAll(imgOutDir, 0755)
	imgs, err := conv.ConvertPDFToImages(pdf1, imgOutDir)
	if err != nil || len(imgs) == 0 {
		t.Fatalf("ConvertPDFToImages failed: %v", err)
	}

	// 10. Test PDF to Word
	docxOut := filepath.Join(tempDir, "converted.docx")
	err = conv.ConvertPDFToWord(pdf1, docxOut)
	if err != nil {
		t.Fatalf("ConvertPDFToWord failed: %v", err)
	}

	// 11. Test Compress PDF
	compressedOut := filepath.Join(tempDir, "compressed.pdf")
	err = conv.CompressPDF(pdf1, compressedOut, "recommended")
	if err != nil {
		t.Fatalf("CompressPDF failed: %v", err)
	}

	t.Log("✅ ALL ENGINE TEST CASES PASSED SUCCESSFULLY!")
}
