package errs

import "errors"

var (
	ErrNotRoot         = errors.New("the command must be run at the root of the module")
	ErrConfigNotFound  = errors.New("configuration file not found")
	ErrReadFailed      = errors.New("failed to read file")
	ErrInvalidFile     = errors.New("invalid file")
	ErrViolationsFound = errors.New("violations found")
)
