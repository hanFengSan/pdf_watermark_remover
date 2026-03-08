package unit_test

import (
	"image"
	"image/color"
	"testing"

	"pdf_watermark_remover/internal/imageproc"
)

func TestApplyReadabilityGuardRemovesIsolatedPixel(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 7, 7))
	fillWhite(img)
	img.SetGray(3, 3, color.Gray{Y: 0})

	out := imageproc.ApplyReadabilityGuard(img)
	if got := out.GrayAt(3, 3).Y; got != 255 {
		t.Fatalf("expected isolated black pixel to be removed, got %d", got)
	}
}

func TestApplyReadabilityGuardPreservesDenseTextLikeBlock(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 20, 20))
	fillWhite(img)

	for y := 8; y <= 12; y++ {
		for x := 8; x <= 12; x++ {
			img.SetGray(x, y, color.Gray{Y: 0})
		}
	}

	out := imageproc.ApplyReadabilityGuard(img)
	black := countBlack(out)
	if black < 20 {
		t.Fatalf("expected dense text-like block to be mostly preserved, black=%d", black)
	}
}

func TestApplyReadabilityGuardRemovesThinContourArtifact(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 100, 100))
	fillWhite(img)

	for x := 10; x <= 70; x++ {
		img.SetGray(x, 20, color.Gray{Y: 0})
		img.SetGray(x, 45, color.Gray{Y: 0})
	}
	for y := 20; y <= 45; y++ {
		img.SetGray(10, y, color.Gray{Y: 0})
		img.SetGray(70, y, color.Gray{Y: 0})
	}

	out := imageproc.ApplyReadabilityGuard(img)
	if black := countBlack(out); black > 8 {
		t.Fatalf("expected thin contour to be removed, black=%d", black)
	}
}

func fillWhite(img *image.Gray) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}
}

func countBlack(img *image.Gray) int {
	b := img.Bounds()
	count := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y == 0 {
				count++
			}
		}
	}
	return count
}
