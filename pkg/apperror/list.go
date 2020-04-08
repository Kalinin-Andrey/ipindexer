package apperror

import(
	"github.com/pkg/errors"
)

// ErrNotFound is error for case when entity not found
var ErrNotFound error = errors.New("Not found")
// ErrBadRequest is error for case when bad request
var ErrBadRequest error = errors.New("Bad request")
// ErrInternal is error for case when smth went wrong
var ErrInternal error = errors.New("Internal error")

