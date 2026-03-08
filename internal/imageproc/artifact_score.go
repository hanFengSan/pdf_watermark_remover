package imageproc

import "image"

const residualDarkThreshold = 0

type ResidualArtifactStats struct {
	FragmentCount   int
	FragmentDensity float64
	Score           float64
}

func AnalyzeResidualArtifacts(img *image.Gray) ResidualArtifactStats {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return ResidualArtifactStats{}
	}

	total := w * h
	artifactArea := 0
	count := 0

	components := CollectConnectedComponents(img, residualDarkThreshold)
	for _, component := range components {
		bw := component.Width()
		bh := component.Height()
		bboxArea := bw * bh
		if bboxArea == 0 || component.Area == 0 {
			continue
		}

		fill := float64(component.Area) / float64(bboxArea)
		maxSide := bw
		if bh > maxSide {
			maxSide = bh
		}
		minSide := bw
		if bh < minSide {
			minSide = bh
		}
		if minSide == 0 {
			continue
		}
		aspect := float64(maxSide) / float64(minSide)

		if component.Area >= 10 && component.Area <= 380 && maxSide >= 6 && maxSide <= 42 && fill >= 0.15 && fill <= 1.0 && aspect <= 6.8 {
			count++
			artifactArea += component.Area
		}
	}

	density := float64(artifactArea) / float64(total)
	countTerm := float64(count) / 60.0
	if countTerm > 1.0 {
		countTerm = 1.0
	}

	score := density*180.0 + countTerm
	return ResidualArtifactStats{FragmentCount: count, FragmentDensity: density, Score: score}
}
