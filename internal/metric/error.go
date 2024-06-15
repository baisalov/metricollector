package metric

import "errors"

var (
	ErrMetricNotFound      = errors.New("metric not found")
	ErrIncorrectMetricType = errors.New("incorrect metric type")
)
