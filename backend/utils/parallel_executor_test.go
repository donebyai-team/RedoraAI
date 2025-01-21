package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestOrderedParallelExecution(t *testing.T) {
	ctx := context.Background()

	errorEquals := func(err error) require.ErrorAssertionFunc {
		return func(t require.TestingT, expected error, msgAndArgs ...interface{}) {
			require.Equal(t, expected, err, msgAndArgs...)
		}
	}

	tests := []struct {
		name        string
		concurrency int
		data        []int
		worker      func(ctx context.Context, item int) (int, error)
		expected    []int
		expectedErr require.ErrorAssertionFunc
	}{
		{
			name:        "empty data",
			concurrency: 2,
			data:        []int{},
			worker:      func(ctx context.Context, item int) (int, error) { return item, nil },
			expected:    nil,
			expectedErr: require.NoError,
		},
		{
			name:        "single item",
			concurrency: 2,
			data:        []int{1},
			worker:      func(ctx context.Context, item int) (int, error) { return item * 2, nil },
			expected:    []int{2},
			expectedErr: require.NoError,
		},
		{
			name:        "multiple items",
			concurrency: 2,
			data:        []int{1, 2, 3, 4},
			worker:      func(ctx context.Context, item int) (int, error) { return item * 2, nil },
			expected:    []int{2, 4, 6, 8},
			expectedErr: require.NoError,
		},
		{
			name:        "concurrency less than data length",
			concurrency: 2,
			data:        []int{1, 2, 3, 4, 5},
			worker: func(ctx context.Context, item int) (int, error) {
				return item * 2, nil
			},
			expected:    []int{2, 4, 6, 8, 10},
			expectedErr: require.NoError,
		},
		{
			name:        "worker error",
			concurrency: 2,
			data:        []int{1, 2, 3, 4},
			worker: func(ctx context.Context, item int) (int, error) {
				if item%2 == 0 {
					return 0, errors.New("even number error")
				}
				return item * 2, nil
			},
			expected:    []int{2, 0, 6, 0},
			expectedErr: errorEquals(multierr.Combine(errors.New("even number error"), errors.New("even number error"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := OrderedParallelExecution(ctx, tt.concurrency, tt.data, tt.worker)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, results)
		})
	}
}
