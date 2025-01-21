package utils

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type job[T any] struct {
	index int
	data  T
}

type executionResult[R any] struct {
	index int
	item  R
	err   error
}

// OrderedParallelExecution executes the worker function concurrently on each item in the data slice.
// The worker function is called with the context and the item to process. The results are returned in
// the same order as the data slice. The concurrency parameter controls how many workers are run
// concurrently.
//
// The err will be nil if no error occurred or it will contain a [multierr.Error] which will contains
// all the errors that occurred during the execution of the worker function but **without** ordering
// of the errors in there.
func OrderedParallelExecution[T any, R any](ctx context.Context, concurrency int, data []T, worker func(ctx context.Context, item T) (R, error)) (results []R, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	if len(data) == 1 {
		result, err := worker(ctx, data[0])
		return []R{result}, err
	}

	if len(data) < concurrency {
		concurrency = len(data)
	}

	jobs := make(chan job[T])
	workerResults := make(chan executionResult[R])

	// Create number of workers (concurrency) to process the work
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				workerOut, workerErr := worker(ctx, job.data)
				workerResults <- executionResult[R]{job.index, workerOut, workerErr}
			}
		}()
	}

	// Feed the work to the workers, wait for them to finish, and close the workerResults channel
	go func() {
		for i, s := range data {
			jobs <- job[T]{i, s}
		}

		close(jobs)

		wg.Wait()
		close(workerResults)
	}()

	// Collect the results from the workers and store them in the results slice ordered
	results = make([]R, len(data))
	for workerResult := range workerResults {
		results[workerResult.index] = workerResult.item
		err = multierr.Append(err, workerResult.err)
	}

	return results, err
}
