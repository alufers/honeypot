package fakeshell

import (
	"errors"
	"fmt"
	"os"
)

var errorMappings = map[error]string{
	os.ErrNotExist:   "No such file or directory",
	os.ErrPermission: "Permission denied",
}

//MapError changes errors to look like UNIX errors
func MapError[R any](v R, err error) (R, error) {
	if err != nil {
		for k, msg := range errorMappings {
			if errors.Is(err, k) {
				return v, fmt.Errorf("%s", msg)
			}
		}
		return v, err
	}
	return v, nil
}
