package pdf

import (
	"fmt"
	"image"
	"os"

	"github.com/phpdave11/gofpdf"

	"pdf_watermark_remover/internal/logutil"
)

func BuildPDFfromImages(imagePaths []string, outputPath string) error {
	if len(imagePaths) == 0 {
		return fmt.Errorf("no images to rebuild PDF")
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt"})
	logutil.Printf("rebuild progress: 0/%d\n", len(imagePaths))
	for _, imgPath := range imagePaths {
		f, err := os.Open(imgPath)
		if err != nil {
			return fmt.Errorf("open image %s: %w", imgPath, err)
		}
		cfg, _, err := image.DecodeConfig(f)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("decode image %s: %w", imgPath, err)
		}

		w, h := float64(cfg.Width), float64(cfg.Height)
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: w, Ht: h})
		opts := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		pdf.ImageOptions(imgPath, 0, 0, w, h, false, opts, 0, "")
		pageNum := pdf.PageNo()
		logutil.Printf("rebuild progress: %d/%d\n", pageNum, len(imagePaths))
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write output pdf: %w", err)
	}
	return nil
}
