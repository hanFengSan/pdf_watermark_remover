package watermark

import (
	"image"
	"image/color"
	"math"
	"sort"
)

func SuppressWatermark(in image.Image, pattern Pattern, bestEffort bool, mode SuppressionMode) *image.Gray {
	b := in.Bounds()
	out := image.NewGray(b)
	gray := toGrayMatrix(in)
	shift := 0
	if bestEffort {
		shift = 8
	}

	strength := 1.0
	if bestEffort {
		strength = 0.72
	}
	switch mode {
	case ModeSafe:
		strength *= 0.78
	case ModeBalanced:
		strength *= 1.0
	case ModeAggressive:
		strength *= 1.28
	}

	minMaskScore := 0.42
	switch mode {
	case ModeSafe:
		minMaskScore = 0.56
	case ModeBalanced:
		minMaskScore = 0.46
	case ModeAggressive:
		minMaskScore = 0.34
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			g8 := gray[y-b.Min.Y][x-b.Min.X]

			tx, ty := templateCoord(x-b.Min.X, y-b.Min.Y, b.Dx(), b.Dy(), pattern.TemplateW, pattern.TemplateH)
			tidx := ty*pattern.TemplateW + tx

			highlight := int(pattern.HighlightSig[tidx])
			bg := int(pattern.Background[tidx])
			thr := int(pattern.Threshold) - shift

			maskScore := watermarkMaskScore(gray, x-b.Min.X, y-b.Min.Y, highlight)

			outGray := int(g8)
			if highlight > 4 && int(g8) >= thr && maskScore >= minMaskScore {
				// Use template-informed subtraction to avoid visible watermark contours.
				delta := int(float64(highlight) * strength)
				candidate := outGray - delta
				bgGuard := bg
				if mode == ModeAggressive {
					bgGuard = bg - 8
				}
				if bgGuard < 0 {
					bgGuard = 0
				}
				if candidate < bgGuard {
					candidate = bgGuard
				}
				if mode == ModeAggressive {
					candidate -= 2
				}
				if candidate < 0 {
					candidate = 0
				}
				if candidate < bgGuard {
					candidate = bg
				}
				outGray = candidate
			}

			if outGray < 0 {
				outGray = 0
			}
			if outGray > 255 {
				outGray = 255
			}
			out.SetGray(x, y, color.Gray{Y: uint8(outGray)})
		}
	}
	return out
}

// SuppressWatermarkSinglePage performs per-page watermark suppression without
// relying on cross-page templates. It is designed for pages with many repeated
// diagonal watermark instances.
func SuppressWatermarkSinglePage(in image.Image, bestEffort bool, mode SuppressionMode) *image.Gray {
	out, _ := SuppressWatermarkSinglePageWithMask(in, bestEffort, mode)
	return out
}

func SuppressWatermarkSinglePageWithMask(in image.Image, bestEffort bool, mode SuppressionMode) (*image.Gray, *image.Gray) {
	b := in.Bounds()
	gray := toGrayMatrix(in)
	localMean := localMeanMatrix(gray, 31)

	type sample struct {
		x, y int
		res  int
		abs  int
		dir  float64
	}

	samples := make([]sample, 0, (b.Dx()*b.Dy())/10)
	for y := 2; y < b.Dy()-2; y++ {
		for x := 2; x < b.Dx()-2; x++ {
			dir := diagonalDirectionScore(gray, x, y)
			if dir < 0.36 {
				continue
			}
			res := int(gray[y][x]) - int(localMean[y][x])
			ar := res
			if ar < 0 {
				ar = -ar
			}
			samples = append(samples, sample{x: x, y: y, res: res, abs: ar, dir: dir})
		}
	}

	if len(samples) == 0 {
		return matrixToGray(gray), emptyMask(b.Dx(), b.Dy())
	}

	absVals := make([]int, len(samples))
	for i := range samples {
		absVals[i] = samples[i].abs
	}
	sort.Ints(absVals)
	lower := absVals[(len(absVals)*70)/100]
	upper := absVals[(len(absVals)*96)/100]
	if upper < lower+2 {
		upper = lower + 2
	}

	blend := 0.84
	switch mode {
	case ModeSafe:
		blend = 0.68
	case ModeBalanced:
		blend = 0.84
	case ModeAggressive:
		blend = 0.93
	}
	if bestEffort {
		blend -= 0.10
	}
	if blend < 0.5 {
		blend = 0.5
	}

	out := make([][]uint8, b.Dy())
	mask := make([][]bool, b.Dy())
	for y := range out {
		out[y] = make([]uint8, b.Dx())
		mask[y] = make([]bool, b.Dx())
		copy(out[y], gray[y])
	}

	for _, s := range samples {
		if s.abs < lower || s.abs > upper {
			continue
		}
		if s.dir < 0.42 {
			continue
		}

		g := float64(gray[s.y][s.x])
		m := float64(localMean[s.y][s.x])
		v := g*(1.0-blend) + m*blend
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		out[s.y][s.x] = uint8(v)
		mask[s.y][s.x] = true
	}

	mask = dilateMask(mask, 1)

	edgeBlend := 0.64
	switch mode {
	case ModeSafe:
		edgeBlend = 0.54
	case ModeBalanced:
		edgeBlend = 0.64
	case ModeAggressive:
		edgeBlend = 0.76
	}
	if bestEffort {
		edgeBlend -= 0.08
	}
	if edgeBlend < 0.45 {
		edgeBlend = 0.45
	}

	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if !mask[y][x] {
				continue
			}
			g := float64(out[y][x])
			m := float64(localMean[y][x])
			v := g*(1.0-edgeBlend) + m*edgeBlend
			if v < 0 {
				v = 0
			}
			if v > 255 {
				v = 255
			}
			out[y][x] = uint8(v)
		}
	}

	return matrixToGray(out), boolMaskToGray(mask)
}

func emptyMask(w, h int) *image.Gray {
	return image.NewGray(image.Rect(0, 0, w, h))
}

func boolMaskToGray(mask [][]bool) *image.Gray {
	h := len(mask)
	if h == 0 {
		return image.NewGray(image.Rect(0, 0, 0, 0))
	}
	w := len(mask[0])
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if mask[y][x] {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	return out
}

func dilateMask(mask [][]bool, radius int) [][]bool {
	h := len(mask)
	if h == 0 {
		return mask
	}
	w := len(mask[0])
	if radius <= 0 {
		return mask
	}

	out := make([][]bool, h)
	for y := range out {
		out[y] = make([]bool, w)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !mask[y][x] {
				continue
			}
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || nx >= w || ny < 0 || ny >= h {
						continue
					}
					out[ny][nx] = true
				}
			}
		}
	}

	return out
}

func toGrayMatrix(img image.Image) [][]uint8 {
	b := img.Bounds()
	h := b.Dy()
	w := b.Dx()
	out := make([][]uint8, h)
	for y := 0; y < h; y++ {
		out[y] = make([]uint8, w)
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			out[y][x] = uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)
		}
	}
	return out
}

func watermarkMaskScore(gray [][]uint8, x, y, highlight int) float64 {
	h := len(gray)
	if h == 0 {
		return 0
	}
	w := len(gray[0])
	if x <= 1 || y <= 1 || x >= w-2 || y >= h-2 {
		return 0
	}

	d45 := math.Abs(float64(gray[y+1][x-1]) - float64(gray[y-1][x+1]))
	d135 := math.Abs(float64(gray[y-1][x-1]) - float64(gray[y+1][x+1]))
	dir := (d45 - d135) / 255.0
	if dir < 0 {
		dir = 0
	}
	if dir > 1 {
		dir = 1
	}

	// local contrast helps suppress random flat-area template artifacts.
	minV := 255
	maxV := 0
	for yy := y - 1; yy <= y+1; yy++ {
		for xx := x - 1; xx <= x+1; xx++ {
			v := int(gray[yy][xx])
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	contrast := float64(maxV-minV) / 255.0
	if contrast > 1 {
		contrast = 1
	}

	templateConf := float64(highlight) / 255.0
	if templateConf > 1 {
		templateConf = 1
	}

	return 0.50*templateConf + 0.35*dir + 0.15*contrast
}

func diagonalDirectionScore(gray [][]uint8, x, y int) float64 {
	d45 := math.Abs(float64(gray[y+1][x-1]) - float64(gray[y-1][x+1]))
	d135 := math.Abs(float64(gray[y-1][x-1]) - float64(gray[y+1][x+1]))
	v := (d45 - d135) / 255.0
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v
}

func localMeanMatrix(gray [][]uint8, window int) [][]uint8 {
	if window < 3 {
		window = 3
	}
	if window%2 == 0 {
		window++
	}
	h := len(gray)
	if h == 0 {
		return nil
	}
	w := len(gray[0])

	integral := make([][]int, h+1)
	for y := range integral {
		integral[y] = make([]int, w+1)
	}
	for y := 1; y <= h; y++ {
		row := 0
		for x := 1; x <= w; x++ {
			row += int(gray[y-1][x-1])
			integral[y][x] = integral[y-1][x] + row
		}
	}

	half := window / 2
	out := make([][]uint8, h)
	for y := 0; y < h; y++ {
		out[y] = make([]uint8, w)
		y1 := y - half
		if y1 < 0 {
			y1 = 0
		}
		y2 := y + half
		if y2 >= h {
			y2 = h - 1
		}
		for x := 0; x < w; x++ {
			x1 := x - half
			if x1 < 0 {
				x1 = 0
			}
			x2 := x + half
			if x2 >= w {
				x2 = w - 1
			}
			sum := integral[y2+1][x2+1] - integral[y1][x2+1] - integral[y2+1][x1] + integral[y1][x1]
			area := (x2 - x1 + 1) * (y2 - y1 + 1)
			out[y][x] = uint8(sum / area)
		}
	}
	return out
}

func matrixToGray(m [][]uint8) *image.Gray {
	h := len(m)
	if h == 0 {
		return image.NewGray(image.Rect(0, 0, 0, 0))
	}
	w := len(m[0])
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.SetGray(x, y, color.Gray{Y: m[y][x]})
		}
	}
	return out
}

func templateCoord(x, y, width, height, tw, th int) (int, int) {
	if width <= 0 || height <= 0 || tw <= 0 || th <= 0 {
		return 0, 0
	}
	tx := (x * tw) / width
	ty := (y * th) / height
	if tx < 0 {
		tx = 0
	}
	if ty < 0 {
		ty = 0
	}
	if tx >= tw {
		tx = tw - 1
	}
	if ty >= th {
		ty = th - 1
	}
	return tx, ty
}
