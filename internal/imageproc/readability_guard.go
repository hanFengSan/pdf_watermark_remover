package imageproc

import (
	"image"
	"image/color"
)

func ApplyReadabilityGuard(img *image.Gray) *image.Gray {
	b := img.Bounds()
	out := image.NewGray(b)
	copy(out.Pix, img.Pix)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if out.GrayAt(x, y).Y != 0 {
				continue
			}
			if blackNeighborCount(out, x, y) <= 1 {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}

	removeContourArtifacts(out)
	return out
}

func removeContourArtifacts(img *image.Gray) {
	components := CollectConnectedComponents(img, 0)
	for _, component := range components {
		bw := component.Width()
		bh := component.Height()
		bboxArea := bw * bh
		if bboxArea == 0 || component.Area == 0 {
			continue
		}
		fillRatio := float64(component.Area) / float64(bboxArea)
		branchRatio := float64(component.Branches) / float64(component.Area)

		if isContourArtifact(component.Area, bw, bh, fillRatio, branchRatio) {
			for _, p := range component.Pixels {
				img.SetGray(p.X, p.Y, color.Gray{Y: 255})
			}
		}
	}
}

func isContourArtifact(area, width, height int, fillRatio, branchRatio float64) bool {
	if area <= 4 {
		return true
	}

	maxSide := width
	if height > maxSide {
		maxSide = height
	}
	minSide := width
	if height < minSide {
		minSide = height
	}
	if minSide <= 0 {
		return false
	}
	aspect := float64(maxSide) / float64(minSide)

	if area <= 140 && maxSide >= 18 && fillRatio <= 0.20 && branchRatio <= 0.12 {
		return true
	}
	if area <= 220 && maxSide >= 24 && aspect >= 2.2 && fillRatio <= 0.16 && branchRatio <= 0.10 {
		return true
	}

	return false
}

func blackNeighborCount(img *image.Gray, x, y int) int {
	b := img.Bounds()
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx < b.Min.X || nx >= b.Max.X || ny < b.Min.Y || ny >= b.Max.Y {
				continue
			}
			if img.GrayAt(nx, ny).Y == 0 {
				count++
			}
		}
	}
	return count
}
