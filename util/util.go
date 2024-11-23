package util

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func Fatalf(format string, args ...interface{}) error {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		_, file := filepath.Split(file)
		return fmt.Errorf("%s:%d: %s", file, line, fmt.Sprintf(format, args...))
	}
	return fmt.Errorf(format, args...)
}

func Fatal(err error) error {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		_, file := filepath.Split(file)
		return fmt.Errorf("%s:%d: %v", file, line, err)
	}
	return err
}
