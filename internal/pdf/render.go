package pdf

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"

	"pdf_watermark_remover/internal/logutil"
)

var (
	pdfiumPoolOnce sync.Once
	pdfiumPool     pdfium.Pool
	pdfiumPoolErr  error
)

func getPDFiumPool() (pdfium.Pool, error) {
	pdfiumPoolOnce.Do(func() {
		poolWorkers := configuredRenderWorkers()
		pdfiumPool, pdfiumPoolErr = webassembly.Init(webassembly.Config{
			MinIdle:      1,
			MaxIdle:      poolWorkers,
			MaxTotal:     poolWorkers,
			ReuseWorkers: true,
		})
	})
	return pdfiumPool, pdfiumPoolErr
}

func RenderToPNGs(inputPDF, outputPrefix string) error {
	pool, err := getPDFiumPool()
	if err != nil {
		return fmt.Errorf("initialize pdfium wasm runtime: %w", err)
	}

	instance, err := pool.GetInstance(60 * time.Second)
	if err != nil {
		return fmt.Errorf("get pdfium instance: %w", err)
	}

	pdfBytes, err := os.ReadFile(inputPDF)
	if err != nil {
		_ = instance.Close()
		return fmt.Errorf("read input pdf: %w", err)
	}

	doc, err := instance.OpenDocument(&requests.OpenDocument{File: &pdfBytes})
	if err != nil {
		_ = instance.Close()
		return fmt.Errorf("open document with pdfium: %w", err)
	}

	pageCountResp, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
		_ = instance.Close()
		return fmt.Errorf("get page count: %w", err)
	}
	if pageCountResp.PageCount == 0 {
		_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
		_ = instance.Close()
		return fmt.Errorf("no pages rendered for %s", inputPDF)
	}

	workers := configuredRenderWorkers()
	if workers > pageCountResp.PageCount {
		workers = pageCountResp.PageCount
	}
	if workers < 1 {
		workers = 1
	}

	progress := newRenderProgress(pageCountResp.PageCount, workers)

	if workers == 1 {
		for i := 0; i < pageCountResp.PageCount; i++ {
			if err := renderPage(instance, doc.Document, i, outputPrefix); err != nil {
				_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
				_ = instance.Close()
				return err
			}
			progress.Tick()
		}
		_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
		_ = instance.Close()
		return nil
	}

	if _, err := instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document}); err != nil {
		return fmt.Errorf("close seed document: %w", err)
	}
	if err := instance.Close(); err != nil {
		return fmt.Errorf("close seed instance: %w", err)
	}

	jobs := make(chan int)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for wi := 0; wi < workers; wi++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			workerInstance, err := pool.GetInstance(60 * time.Second)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("get pdfium worker instance: %w", err):
				default:
				}
				return
			}
			defer workerInstance.Close()

			workerDoc, err := workerInstance.OpenDocument(&requests.OpenDocument{File: &pdfBytes})
			if err != nil {
				select {
				case errCh <- fmt.Errorf("open worker document: %w", err):
				default:
				}
				return
			}
			defer workerInstance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: workerDoc.Document})

			for pageIndex := range jobs {
				if err := renderPage(workerInstance, workerDoc.Document, pageIndex, outputPrefix); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				progress.Tick()
			}
		}()
	}

	for i := 0; i < pageCountResp.PageCount; i++ {
		select {
		case jobs <- i:
		case err := <-errCh:
			close(jobs)
			wg.Wait()
			return err
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	return nil
}

func renderPage(instance pdfium.Pdfium, document references.FPDF_DOCUMENT, pageIndex int, outputPrefix string) error {
	pageRender, err := instance.RenderPageInDPI(&requests.RenderPageInDPI{
		DPI: 400,
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{Document: document, Index: pageIndex},
		},
	})
	if err != nil {
		return fmt.Errorf("render page %d: %w", pageIndex+1, err)
	}
	defer pageRender.Cleanup()

	outputPath := fmt.Sprintf("%s-%04d.png", outputPrefix, pageIndex+1)
	if err := writePNG(outputPath, pageRender.Result.Image); err != nil {
		return fmt.Errorf("write rendered page %d: %w", pageIndex+1, err)
	}
	return nil
}

func configuredRenderWorkers() int {
	cpuCount := runtime.NumCPU()
	if cpuCount < 1 {
		cpuCount = 1
	}

	if raw := os.Getenv("WM_RENDER_WORKERS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			if n > cpuCount {
				return cpuCount
			}
			return n
		}
	}

	if cpuCount >= 4 {
		return 4
	}

	targetPct := 80
	if raw := os.Getenv("WM_RENDER_CPU_TARGET"); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil {
			if p < 10 {
				p = 10
			}
			if p > 100 {
				p = 100
			}
			targetPct = p
		}
	}

	w := (cpuCount*targetPct + 99) / 100
	if w < 1 {
		w = 1
	}
	if w > cpuCount {
		w = cpuCount
	}
	return w
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

type renderProgress struct {
	total int32
	done  atomic.Int32
}

func newRenderProgress(total, workers int) *renderProgress {
	r := &renderProgress{total: int32(total)}
	logutil.Printf("rendering pages: 0/%d (workers=%d)\n", total, workers)
	return r
}

func (r *renderProgress) Tick() {
	current := r.done.Add(1)
	logutil.Printf("rendered pages: %d/%d\n", current, r.total)
}
