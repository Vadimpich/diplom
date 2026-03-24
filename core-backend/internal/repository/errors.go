package repository

import "errors"

var ErrNotFound = errors.New("repository: not found")
var ErrConflict = errors.New("repository: conflict")
