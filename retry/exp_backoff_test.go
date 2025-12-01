//go:build integrationtest || !unittest

package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simplesurance/cfdns/log"
	"github.com/simplesurance/cfdns/log/testtarget"
	"github.com/simplesurance/cfdns/retry"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	someErr := errors.New("some error")

	cases := []*struct {
		name              string
		attempts          int
		tempErrorCount    int
		permErrorAt       int
		wantError         bool
		wantFunctionCalls int
	}{
		{
			name:              "Max1AttemptSuccess",
			attempts:          1,
			wantFunctionCalls: 1,
		},
		{
			name:              "Max1AttemptTemporaryError",
			attempts:          1,
			tempErrorCount:    10,
			wantError:         true,
			wantFunctionCalls: 1,
		},
		{
			name:              "Max2AttemptsSucceedOn2",
			attempts:          2,
			tempErrorCount:    1,
			wantError:         false,
			wantFunctionCalls: 2,
		},
		{
			name:              "Max10AttemptsPermErrorOn4",
			attempts:          10,
			tempErrorCount:    5,
			permErrorAt:       6,
			wantError:         true,
			wantFunctionCalls: 6,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.New(testtarget.ForTest(t, true),
				log.WithDebugEnabledFn(func() bool { return true }))
			fCallCount := 0

			err := retry.ExpBackoff(
				ctx,
				logger,
				time.Millisecond,
				10*time.Millisecond,
				2,
				tc.attempts,
				func() error {
					fCallCount++

					if tc.permErrorAt != 0 && fCallCount == tc.permErrorAt {
						t.Logf("test function returning permanent error (c=%d l=%d)",
							fCallCount, tc.permErrorAt)

						return retry.PermanentError{Cause: someErr}
					}

					if tc.tempErrorCount != 0 && fCallCount <= tc.tempErrorCount {
						t.Logf("test function returning temporary error (c=%d l=%d)",
							fCallCount, tc.tempErrorCount)

						return someErr
					}

					t.Logf("test function returning nil (c=%d pl=%d tl=%d)",
						fCallCount, tc.permErrorAt, tc.tempErrorCount)

					return nil
				})
			if tc.wantError {
				assertEquals(t, someErr, err)
			} else {
				assertNoError(t, err)
			}

			assertEquals(t, tc.wantFunctionCalls, fCallCount)
		})
	}
}

func TestContextCancel(t *testing.T) {
	logger := log.New(testtarget.ForTest(t, true),
		log.WithDebugEnabledFn(func() bool { return true }))

	ctx, done := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer done()

	err := retry.ExpBackoff(ctx, logger, time.Hour, time.Hour, 2, 100, func() error {
		return errors.New("some app error")
	})

	assertErrorIs(t, err, context.DeadlineExceeded)
}

func assertEquals(t *testing.T, v1, v2 any) {
	if v1 != v2 {
		t.Errorf("want: %v, have %v", v1, v2)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func assertErrorIs(t *testing.T, err, target error) {
	if !errors.Is(err, target) {
		t.Errorf("Want error %T, got %v", target, err)
	}
}
