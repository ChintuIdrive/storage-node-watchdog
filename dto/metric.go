package dto

import "time"

type Metric[T any] struct {
	Name             string
	Value            T
	Threshold        T
	LastAlertTime    time.Time
	HighLoadDuration time.Duration //Represents how long the value has been above the threshold
}
