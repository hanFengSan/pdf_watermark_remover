package pipeline

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"time"

	"pdf_watermark_remover/internal/imageproc"
	"pdf_watermark_remover/internal/logutil"
	"pdf_watermark_remover/internal/memutil"
	"pdf_watermark_remover/internal/output"
	pdfops "pdf_watermark_remover/internal/pdf"
	"pdf_watermark_remover/internal/watermark"
)

type Runner struct{}

type Result struct {
	InputPath  string
	OutputPath string
	PageCount  int
	Duration   time.Duration
}

type pageResult struct {
	final    image.Image
	mask     *image.Gray
	artifact imageproc.ResidualArtifactStats
}

func NewRunner() *Runner { return &Runner{} }

func (r *Runner) Run(ctx context.Context, inputPath string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: fmt.Sprintf("context canceled: %v", err)}
	}
	start := time.Now()
	tuning := DefaultTuningConfig()
	logutil.Println("phase: validating input")

	info, err := ValidateInput(inputPath)
	if err != nil {
		return Result{}, err
	}
	logutil.Printf("phase: input validated (pages=%d)\n", info.PageCount)

	outputPath, err := output.NextOutputPath(info.Path)
	if err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: err.Error()}
	}

	workDir, err := os.MkdirTemp("", "pdf-watermark-remover-*")
	if err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: fmt.Sprintf("create temp dir: %v", err)}
	}
	defer os.RemoveAll(workDir)

	renderPrefix := filepath.Join(workDir, "page")
	logutil.Println("phase: rendering pdf pages")
	if err := pdfops.RenderToPNGs(info.Path, renderPrefix); err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: err.Error()}
	}
	memutil.ForceRelease()

	pagePaths, err := filepath.Glob(renderPrefix + "-*.png")
	if err != nil || len(pagePaths) == 0 {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: "failed to gather rendered pages"}
	}
	sort.Strings(pagePaths)

	pattern, err := watermark.EstimatePattern(pagePaths)
	if err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: fmt.Sprintf("estimate watermark: %v", err)}
	}
	logutil.Println("phase: watermark pattern estimated")
	bestEffort := watermark.SelectBestEffortMode(pattern)

	clusterByPath, groups, err := watermark.AssignClusters(pagePaths)
	if err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: fmt.Sprintf("cluster pages: %v", err)}
	}

	patternByCluster := map[int]watermark.Pattern{}
	for clusterID, paths := range groups {
		if len(paths) == 0 {
			continue
		}
		cp, err := watermark.EstimatePattern(paths)
		if err != nil {
			continue
		}
		patternByCluster[clusterID] = cp
	}

	processedPaths := make([]string, len(pagePaths))
	for i := range pagePaths {
		processedPaths[i] = filepath.Join(workDir, fmt.Sprintf("processed-%04d.png", i+1))
	}

	wmMode := os.Getenv("WM_MODE")
	if wmMode == "" {
		wmMode = "single"
	}
	debugMask := os.Getenv("WM_DEBUG_MASK") == "1"
	debugPage := os.Getenv("WM_DEBUG_PAGE") == "1"

	progress := NewProgress(len(pagePaths))
	trimEvery := 10
	logutil.Println("phase: processing pages")
	err = processInParallel(pagePaths, func(path string) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context canceled: %w", err)
		}
		idx := indexOf(pagePaths, path)
		img, err := loadPNG(path)
		if err != nil {
			return err
		}
		activePattern := selectPatternForPage(path, pattern, clusterByPath, patternByCluster)

		profile := imageproc.AnalyzePage(img)
		mode := watermark.SelectSuppressionMode(activePattern, profile.DarkRatio, profile.StdDev)

		base := processPageImage(img, activePattern, bestEffort, mode, false, wmMode, tuning)
		chosen := base

		if shouldFallback(base.artifact, tuning) {
			usedFallback := false
			retryMode := mode
			if retryMode != watermark.ModeAggressive {
				retryMode = watermark.ModeAggressive
			}
			retry := processPageImage(img, activePattern, false, retryMode, true, wmMode, tuning)
			if improvedEnough(base.artifact, retry.artifact, tuning) {
				chosen = retry
				usedFallback = true
			}
			if debugPage {
				logutil.Printf("page=%d fallback=%t base(score=%.3f,count=%d) retry(score=%.3f,count=%d)\n",
					idx+1,
					usedFallback,
					base.artifact.Score,
					base.artifact.FragmentCount,
					retry.artifact.Score,
					retry.artifact.FragmentCount,
				)
			}
		} else if debugPage {
			logutil.Printf("page=%d fallback=false base(score=%.3f,count=%d)\n", idx+1, base.artifact.Score, base.artifact.FragmentCount)
		}

		if debugMask && chosen.mask != nil {
			_ = savePNG(filepath.Join(workDir, fmt.Sprintf("mask-%04d.png", idx+1)), chosen.mask)
		}

		if err := savePNG(processedPaths[idx], chosen.final); err != nil {
			return err
		}
		done := progress.Tick()
		if done%trimEvery == 0 {
			memutil.ForceRelease()
		}
		return nil
	})
	if err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: fmt.Sprintf("process pages: %v", err)}
	}
	memutil.ForceRelease()

	logutil.Println("phase: rebuilding output pdf")
	if err := pdfops.BuildPDFfromImages(processedPaths, outputPath); err != nil {
		return Result{}, &ExitError{Code: ExitProcessing, Msg: err.Error()}
	}
	logutil.Println("phase: rebuild complete")
	memutil.ForceRelease()

	_ = progress.Duration()
	return Result{InputPath: info.Path, OutputPath: outputPath, PageCount: len(processedPaths), Duration: time.Since(start)}, nil
}

func selectPatternForPage(path string, defaultPattern watermark.Pattern, clusterByPath map[string]int, patternByCluster map[int]watermark.Pattern) watermark.Pattern {
	activePattern := defaultPattern
	if cid, ok := clusterByPath[path]; ok {
		if cp, ok := patternByCluster[cid]; ok {
			activePattern = cp
		}
	}
	return activePattern
}

func processPageImage(img image.Image, activePattern watermark.Pattern, bestEffort bool, mode watermark.SuppressionMode, aggressiveRetry bool, wmMode string, tuning TuningConfig) pageResult {
	var suppressed *image.Gray
	var wmMask *image.Gray
	switch wmMode {
	case "hybrid":
		suppressed = watermark.SuppressWatermark(img, activePattern, bestEffort, mode)
	default:
		suppressed, wmMask = watermark.SuppressWatermarkSinglePageWithMask(img, bestEffort, mode)
	}

	postProfile := imageproc.AnalyzePage(suppressed)
	if postProfile.CoverLike {
		cover := imageproc.EnhanceCover(suppressed)
		return pageResult{final: cover, mask: wmMask}
	}

	thr := int(imageproc.ComputeAdaptiveThreshold(suppressed, postProfile))
	window := tuning.BaseWindowSize
	sensitivity := tuning.BaseSensitivity
	if aggressiveRetry {
		thr += tuning.RetryThresholdBoost
		if thr > tuning.RetryThresholdMax {
			thr = tuning.RetryThresholdMax
		}
		window = tuning.RetryWindowSize
		sensitivity = tuning.RetrySensitivity
	}

	bin := imageproc.LocalAdaptiveBinarizeWithMask(suppressed, wmMask, uint8(thr), window, sensitivity)
	bin = imageproc.ApplyReadabilityGuard(bin)
	stats := imageproc.AnalyzeResidualArtifacts(bin)
	return pageResult{final: bin, mask: wmMask, artifact: stats}
}

func shouldFallback(base imageproc.ResidualArtifactStats, tuning TuningConfig) bool {
	return base.Score > tuning.FallbackMinScore && base.FragmentCount >= tuning.FallbackMinFragments
}

func improvedEnough(base, retry imageproc.ResidualArtifactStats, tuning TuningConfig) bool {
	return retry.Score+tuning.FallbackMinImprovement < base.Score
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func indexOf(items []string, item string) int {
	for i := range items {
		if items[i] == item {
			return i
		}
	}
	return 0
}
