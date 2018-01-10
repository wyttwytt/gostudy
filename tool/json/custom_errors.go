package json

import (
	"errors"
)

var (
	InvalidJson      = errors.New("tool.json: provided json data is invalid")
	JsonPathNotFound = errors.New("tool.json: json path not found")
)
