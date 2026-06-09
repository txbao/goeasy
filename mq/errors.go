package mq

import "errors"

var ErrInvalidPayload = errors.New("mq: payload is not valid JSON")
