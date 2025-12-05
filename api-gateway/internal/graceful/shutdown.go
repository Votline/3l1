package graceful

import (
	"context"
)

func Shutdown(stop func() error, ctx context.Context) error {
	done := make(chan error, 1)

	go func(){
		done <-stop()
	}()

	select{
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
