package utils

import (
	"context"
	"errors"
	"time"
)

func Retry(
	ctx context.Context,
	callback func() error,
	interval time.Duration,
	ignoredErrs ...error,
) error {
	var (
		ticker     *time.Ticker = time.NewTicker(interval)
		ignoredErr error        = errors.Join(ignoredErrs...)
		err        error        = nil
	)

	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()

		case <-ticker.C:
			err = callback()
			if err == nil || errors.Is(err, ignoredErr) {
				return nil
			}
		}
	}
}
