package unit_test

import (
	"image"
	"image/color"
	"testing"

	"pdf_watermark_remover/internal/imageproc"
)

func TestAnalyzeResidualArtifactsLowForCleanPage(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 200, 200))
	fillWhite(img)

	for y := 70; y < 130; y++ {
		for x := 40; x < 160; x++ {
			if (x+y)%9 == 0 {
				img.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	stats := imageproc.AnalyzeResidualArtifacts(img)
	if stats.Score > 0.9 {
		t.Fatalf("expected low residual artifact score, got %.3f", stats.Score)
	}
}

func TestAnalyzeResidualArtifactsHighForFragmentedResiduals(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 220, 220))
	fillWhite(img)

	for r := 0; r < 28; r++ {
		baseX := 10 + (r%7)*28
		baseY := 10 + (r/7)*48
		for y := baseY; y < baseY+5; y++ {
			for x := baseX; x < baseX+9; x++ {
				img.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	stats := imageproc.AnalyzeResidualArtifacts(img)
	if stats.FragmentCount < 20 {
		t.Fatalf("expected many residual fragments, got %d", stats.FragmentCount)
	}
	if stats.Score <= 0.95 {
		t.Fatalf("expected high residual artifact score, got %.3f", stats.Score)
	}
}
