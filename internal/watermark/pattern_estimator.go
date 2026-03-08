package watermark

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"sort"
)

func AssignClusters(imagePaths []string) (map[string]int, map[int][]string, error) {
	assignments := make(map[string]int, len(imagePaths))
	groups := map[int][]string{0: {}, 1: {}}

	for _, p := range imagePaths {
		mean, darkRatio, err := quickPageStats(p)
		if err != nil {
			return nil, nil, err
		}
		cluster := 0
		if mean < 150 && darkRatio > 0.35 {
			cluster = 1
		}
		assignments[p] = cluster
		groups[cluster] = append(groups[cluster], p)
	}

	if len(groups[0]) == 0 || len(groups[1]) == 0 {
		return assignments, map[int][]string{0: imagePaths}, nil
	}

	return assignments, groups, nil
}

type Pattern struct {
	Threshold    uint8
	Confidence   float64
	TemplateW    int
	TemplateH    int
	Background   []uint8
	HighlightSig []uint8
}

func EstimatePattern(imagePaths []string) (Pattern, error) {
	if len(imagePaths) == 0 {
		return Pattern{}, fmt.Errorf("no pages for pattern estimation")
	}

	const templateW = 192
	const templateH = 192
	cells := templateW * templateH
	pageCount := len(imagePaths)
	gridValues := make([]uint8, cells*pageCount)

	var total uint64
	var count uint64
	for pageIdx, p := range imagePaths {
		f, err := os.Open(p)
		if err != nil {
			return Pattern{}, err
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			return Pattern{}, err
		}
		b := img.Bounds()
		stepX := max(1, b.Dx()/200)
		stepY := max(1, b.Dy()/200)
		for y := b.Min.Y; y < b.Max.Y; y += stepY {
			for x := b.Min.X; x < b.Max.X; x += stepX {
				r, g, b, _ := img.At(x, y).RGBA()
				gray := uint8(((r>>8)+(g>>8)+(b>>8))/3)
				total += uint64(gray)
				count++
			}
		}

		for ty := 0; ty < templateH; ty++ {
			sy := b.Min.Y + (ty*b.Dy())/templateH
			for tx := 0; tx < templateW; tx++ {
				sx := b.Min.X + (tx*b.Dx())/templateW
				r, g, bl, _ := img.At(sx, sy).RGBA()
				gray := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
				idx := ty*templateW + tx
				gridValues[pageIdx*cells+idx] = gray
			}
		}
	}
	mean := uint8(total / count)
	thr := uint8(min(245, int(mean)+28))

	background := make([]uint8, cells)
	highlightSig := make([]uint8, cells)

	var signalSum float64
	for idx := 0; idx < cells; idx++ {
		vals := make([]int, pageCount)
		for p := 0; p < pageCount; p++ {
			vals[p] = int(gridValues[p*cells+idx])
		}
		sort.Ints(vals)
		median := vals[pageCount/2]
		p85i := int(float64(pageCount-1) * 0.85)
		p85 := vals[p85i]
		signal := max(0, p85-median)

		background[idx] = uint8(median)
		highlightSig[idx] = uint8(min(255, signal))
		signalSum += float64(signal)
	}

	avgSignal := signalSum / float64(cells)
	confidence := 0.35 + minFloat(0.55, avgSignal/18.0)
	if len(imagePaths) < 4 {
		confidence *= 0.9
	}
	if confidence > 0.95 {
		confidence = 0.95
	}
	if confidence < 0.25 {
		confidence = 0.25
	}

	return Pattern{
		Threshold:    thr,
		Confidence:   confidence,
		TemplateW:    templateW,
		TemplateH:    templateH,
		Background:   background,
		HighlightSig: highlightSig,
	}, nil
}

func quickPageStats(path string) (float64, float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	img, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return 0, 0, err
	}

	b := img.Bounds()
	stepX := max(1, b.Dx()/220)
	stepY := max(1, b.Dy()/220)

	var count uint64
	var sum uint64
	var dark uint64
	for y := b.Min.Y; y < b.Max.Y; y += stepY {
		for x := b.Min.X; x < b.Max.X; x += stepX {
			r, g, bl, _ := img.At(x, y).RGBA()
			gray := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
			sum += uint64(gray)
			if gray < 96 {
				dark++
			}
			count++
		}
	}

	if count == 0 {
		return 0, 0, nil
	}

	mean := float64(sum) / float64(count)
	darkRatio := float64(dark) / float64(count)
	return mean, darkRatio, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
