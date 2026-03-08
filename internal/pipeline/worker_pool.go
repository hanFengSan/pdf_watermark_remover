package pipeline

import (
	"os"
	"runtime"
	"strconv"
	"sync"
)

func processInParallel[T any](items []T, fn func(T) error) error {
	workers := configuredProcessWorkers()
	if workers < 2 {
		workers = 2
	}
	if workers > len(items) {
		workers = len(items)
	}
	if workers <= 0 {
		return nil
	}

	jobs := make(chan T)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				if err := fn(item); err != nil {
					select {
					case errCh <- err:
					default:
					}
				}
			}
		}()
	}

	for _, item := range items {
		jobs <- item
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func configuredProcessWorkers() int {
	if raw := os.Getenv("WM_PROCESS_WORKER"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			maxCPU := runtime.NumCPU()
			if maxCPU < 1 {
				maxCPU = 1
			}
			if n > maxCPU {
				return maxCPU
			}
			return n
		}
	}

	maxCPU := runtime.NumCPU()
	if maxCPU < 1 {
		maxCPU = 1
	}
	if maxCPU >= 4 {
		return 4
	}
	return maxCPU
}
