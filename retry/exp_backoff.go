package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/simplesurance/cfdns/log"
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
// maxTries indicates how many times the function is invoked. The value
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
	logger *log.Logger,
	firstDelay, maxDelay time.Duration,
	factor float64,
	maxTries int,
	f func() error,
) error {
	start := time.Now()
	delay := firstDelay.Seconds()

	for attempt := 1; ; attempt++ {
		err := f()
		if err == nil {
			return nil
		}

		if maxTries > 0 && attempt >= maxTries {
			logger.W("f() kept failing, exhausting retry limit; giving up.",
				log.WithInt("attempt", attempt),
				log.WithDuration("total_delay", time.Since(start)),
				log.WithError(err))

			var permError PermanentError
			if errors.As(err, &permError) {
				return permError.Cause
			}

			return err
		}

		var permError PermanentError
		if errors.As(err, &permError) {
			logger.W("f() returned a permanent error; giving up.",
				log.WithInt("attempt", attempt),
				log.WithDuration("total_delay", time.Since(start)),
				log.WithError(permError.Cause))

			return permError.Cause
		}

		logger.W("f() returned an error",
			log.WithInt("attempt", attempt),
			log.WithDuration("total_delay", time.Since(start)),
			log.WithError(err))

		select {
		case <-ctx.Done():
			err := ctx.Err()
			logger.D(func(lg log.DebugFn) {
				lg("context was canceled",
					log.WithInt("attempt", attempt),
					log.WithDuration("total_delay", time.Since(start)),
					log.WithError(err))
			})
			return err
		case <-time.After(time.Duration(delay * float64(time.Second))):
			delay *= factor
			logger.D(func(lg log.DebugFn) {
				lg(fmt.Sprintf("next delay: %f seconds", delay))
			})
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
