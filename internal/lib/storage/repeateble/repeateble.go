package repeateble

import (
	"errors"
	"time"
)

func DoWithTries(fn func() error, attemps int, delay time.Duration) error {
	if attemps == 0 {
		return errors.New("attemps count can not be 0")
	}

	var err error

	for range attemps {
		if err = fn(); err != nil {
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return err
}
