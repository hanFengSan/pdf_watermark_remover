package imageproc

import (
	"image"
	"image/color"
	"math"
)

type PageProfile struct {
	Mean      float64
	StdDev    float64
	DarkRatio float64
	CoverLike bool
}

func AnalyzePage(img image.Image) PageProfile {
	h := histogram(img)
	total := float64(pixelCount(img))
	if total == 0 {
		return PageProfile{}
	}

	var sum float64
	var dark float64
	for i := 0; i < 256; i++ {
		c := float64(h[i])
		sum += float64(i) * c
		if i < 96 {
			dark += c
		}
	}
	mean := sum / total

	var varAcc float64
	for i := 0; i < 256; i++ {
		c := float64(h[i])
		d := float64(i) - mean
		varAcc += d * d * c
	}
	std := math.Sqrt(varAcc / total)
	darkRatio := dark / total

	coverLike := mean < 140 && darkRatio > 0.45 && std < 70

	return PageProfile{Mean: mean, StdDev: std, DarkRatio: darkRatio, CoverLike: coverLike}
}

func ComputeAdaptiveThreshold(img image.Image, profile PageProfile) uint8 {
	base := otsuThreshold(img)

	// Bias text pages slightly toward lighter background to reduce stroke adhesion.
	bias := 10
	if profile.DarkRatio > 0.25 {
		bias = 6
	}
	thr := int(base) + bias
	if thr < 120 {
		thr = 120
	}
	if thr > 210 {
		thr = 210
	}
	return uint8(thr)
}

func EnhanceCover(img image.Image) *image.Gray {
	h := histogram(img)
	p5 := percentileFromHistogram(h, 0.05)
	p95 := percentileFromHistogram(h, 0.95)
	if p95 <= p5 {
		p95 = p5 + 1
	}

	b := img.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			gray := float64(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
			norm := (gray - float64(p5)) / float64(p95-p5)
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			// gamma < 1 brightens dark areas while preserving details.
			lifted := math.Pow(norm, 0.85)
			out.SetGray(x, y, color.Gray{Y: uint8(lifted * 255)})
		}
	}
	return out
}

func Binarize(img image.Image, threshold uint8) *image.Gray {
	b := img.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			gray := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
			if gray >= threshold {
				out.SetGray(x, y, color.Gray{Y: 255})
			} else {
				out.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return out
}

func LocalAdaptiveBinarize(img image.Image, baseThreshold uint8, window int, sensitivity float64) *image.Gray {
	return LocalAdaptiveBinarizeWithMask(img, nil, baseThreshold, window, sensitivity)
}

func LocalAdaptiveBinarizeWithMask(img image.Image, mask *image.Gray, baseThreshold uint8, window int, sensitivity float64) *image.Gray {
	if window < 3 {
		window = 3
	}
	if window%2 == 0 {
		window++
	}
	if sensitivity < 0 {
		sensitivity = 0
	}
	if sensitivity > 0.5 {
		sensitivity = 0.5
	}

	b := img.Bounds()
	gray := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			g8 := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
			gray.SetGray(x, y, color.Gray{Y: g8})
		}
	}

	integral := buildIntegral(gray)
	winHalf := window / 2
	base := int(baseThreshold)
	out := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		y1 := maxInt(b.Min.Y, y-winHalf)
		y2 := minInt(b.Max.Y-1, y+winHalf)
		for x := b.Min.X; x < b.Max.X; x++ {
			x1 := maxInt(b.Min.X, x-winHalf)
			x2 := minInt(b.Max.X-1, x+winHalf)
			area := (x2 - x1 + 1) * (y2 - y1 + 1)
			localMean := sumRect(integral, x1-b.Min.X, y1-b.Min.Y, x2-b.Min.X, y2-b.Min.Y) / area

			thr := int(float64(localMean) * (1.0 - sensitivity))
			if thr < base-20 {
				thr = base - 20
			}
			if thr > base+25 {
				thr = base + 25
			}

			if mask != nil && mask.Bounds() == b {
				if mask.GrayAt(x, y).Y > 0 {
					thr += 16
				}
			}
			if thr > 245 {
				thr = 245
			}

			gval := int(gray.GrayAt(x, y).Y)
			if gval >= thr {
				out.SetGray(x, y, color.Gray{Y: 255})
			} else {
				out.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	return out
}

func histogram(img image.Image) [256]uint64 {
	var h [256]uint64
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			gray := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
			h[gray]++
		}
	}
	return h
}

func pixelCount(img image.Image) int {
	b := img.Bounds()
	return b.Dx() * b.Dy()
}

func percentileFromHistogram(h [256]uint64, p float64) int {
	var total uint64
	for i := 0; i < 256; i++ {
		total += h[i]
	}
	target := uint64(float64(total) * p)
	if target == 0 {
		target = 1
	}
	var acc uint64
	for i := 0; i < 256; i++ {
		acc += h[i]
		if acc >= target {
			return i
		}
	}
	return 255
}

func otsuThreshold(img image.Image) uint8 {
	h := histogram(img)
	total := float64(pixelCount(img))
	if total == 0 {
		return 160
	}

	var sumAll float64
	for i := 0; i < 256; i++ {
		sumAll += float64(i) * float64(h[i])
	}

	var sumB float64
	var wB float64
	maxVar := -1.0
	best := 160

	for t := 0; t < 256; t++ {
		wB += float64(h[t])
		if wB == 0 {
			continue
		}
		wF := total - wB
		if wF == 0 {
			break
		}
		sumB += float64(t) * float64(h[t])
		mB := sumB / wB
		mF := (sumAll - sumB) / wF
		between := wB * wF * (mB - mF) * (mB - mF)
		if between > maxVar {
			maxVar = between
			best = t
		}
	}

	return uint8(best)
}

func buildIntegral(img *image.Gray) [][]int {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	integral := make([][]int, h+1)
	for y := range integral {
		integral[y] = make([]int, w+1)
	}
	for y := 1; y <= h; y++ {
		row := 0
		for x := 1; x <= w; x++ {
			row += int(img.GrayAt(b.Min.X+x-1, b.Min.Y+y-1).Y)
			integral[y][x] = integral[y-1][x] + row
		}
	}
	return integral
}

func sumRect(integral [][]int, x1, y1, x2, y2 int) int {
	return integral[y2+1][x2+1] - integral[y1][x2+1] - integral[y2+1][x1] + integral[y1][x1]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
