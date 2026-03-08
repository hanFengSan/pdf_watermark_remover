package unit_test

import (
	"image"
	"image/color"
	"testing"

	"pdf_watermark_remover/internal/imageproc"
)

func TestBinarizeThreshold(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 2, 1))
	img.SetGray(0, 0, color.Gray{Y: 120})
	img.SetGray(1, 0, color.Gray{Y: 220})
	out := imageproc.Binarize(img, 160)
	if got := out.GrayAt(0, 0).Y; got != 0 {
		t.Fatalf("expected black pixel, got %d", got)
	}
	if got := out.GrayAt(1, 0).Y; got != 255 {
		t.Fatalf("expected white pixel, got %d", got)
	}
}
