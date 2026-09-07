package runner

import "sync"

// Map applies fn to each item with bounded parallelism, preserving input
// order in the returned slice regardless of completion order. If onDone is
// provided, it is invoked after each item completes.
func Map[T any, R any](items []T, parallelism int, fn func(T) R, onDone ...func()) []R {
	if parallelism < 1 {
		parallelism = 1
	}
	if parallelism > len(items) {
		parallelism = len(items)
	}
	results := make([]R, len(items))
	if parallelism == 0 {
		return results
	}
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	for i, it := range items {
		wg.Add(1)
		go func(i int, it T) {
			defer wg.Done()
			sem <- struct{}{}
			results[i] = fn(it)
			<-sem
			if len(onDone) > 0 {
				onDone[0]()
			}
		}(i, it)
	}
	wg.Wait()
	return results
}
