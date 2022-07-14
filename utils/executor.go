package utils

import "time"

// Executor executes a function at a specified interval (exactly)
// and returns a channel which may be sent to in order to cancel the executor.
func Executor(interval time.Duration, run func()) chan struct{} {
	stop := make(chan struct{})
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				run()
			case <-stop:
				close(stop)
				ticker.Stop()
				return
			}
		}
	}()
	return stop
}
