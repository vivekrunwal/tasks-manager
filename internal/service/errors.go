package service

import "errors"

var (
    ErrNotFound        = errors.New("task not found")
    ErrVersionConflict = errors.New("version conflict")
)


