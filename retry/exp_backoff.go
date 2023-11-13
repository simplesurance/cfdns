package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/simplesurance/cfdns/logs"
)

// ExpBackoff executes the provided function, retrying it with exponential
// back-off if it fails. Context cancellation is respected. The callback
// function can mark an error as permanent by wrapping it will
// retry.PermanentError, in which case no more attempts will be made.
//
// firstDelay specifies how long it should wait before trying again before
// retrying the first time. On each retry the delay will be multiplied by
// the provided factor, but will not be longer than maxDelay.
//
// attempts indicates how many times the function call be invoked. The value
// 1 means to call it only once, never retrying if it fails. The number 2
// allows for 1 retry, and so on. A value of 0 or less will make it retry
// forever. If f() keeps failing after the number of retries is reached it
// will return the last result of f().
//
// If the provided context expires or is canceled the context error will be
// returned immediately.
//
// f is the function that will be executed. If it returns nil, this function
// will immediately return nil. If it returns a PermanentError
// no new attempt to retry will be executed. The error wrapped by it will be
// returned. If other error is returned, the delay logic will be executed.
func ExpBackoff(
	ctx context.Context,
	logger logs.Driver,
	firstDelay, maxDelay time.Duration,
	factor float64,
	attempts int,
	f func() error,
) error {
	start := time.Now()
	delay := firstDelay.Seconds()

	for attempt := 1; ; attempt++ {
		err := f()
		if err == nil {
			return nil
		}

		if attempts > 0 && attempt >= attempts {
			sisulog.Debug(ctx,
				fmt.Sprintf("f() kept failing after %d attempts and %v; giving up.",
					attempt,
					time.Since(start)),
				logkeys.Error, err)

			var permError PermanentError
			if errors.As(err, &permError) {
				return permError.Cause
			}

			return err
		}

		var permError PermanentError
		if errors.As(err, &permError) {
			sisulog.Debug(ctx,
				fmt.Sprintf("f() returned a permanent error after %d attempts and %v; giving up.",
					attempt,
					time.Since(start)),
				logkeys.Error, permError.Cause)

			return permError.Cause
		}

		sisulog.Debug(ctx,
			fmt.Sprintf("f() returned an error after %d attempts and %v",
				attempt,
				time.Since(start)),
			logkeys.Error, err)

		select {
		case <-ctx.Done():
			err := ctx.Err()
			sisulog.Debug(ctx,
				fmt.Sprintf("context was canceled after %d attempts and %v",
					attempt,
					time.Since(start)),
				logkeys.Error, err)
			return err
		case <-time.After(time.Duration(delay * float64(time.Second))):
			delay *= factor
			if delay > maxDelay.Seconds() {
				delay = maxDelay.Seconds()
			}
		}
	}
}

type PermanentError struct {
	Cause error
}

func (p PermanentError) Unwrap() error {
	return p.Cause
}

func (p PermanentError) Error() string {
	return fmt.Sprintf("permanent error: %v", p.Cause)
}

var _ error = PermanentError{}
