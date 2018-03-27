package dto

import (
	"encoding/gob"
	"time"
)

// SensorMessage is a representation of a name and value.
type SensorMessage struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

func init() {
	gob.Register(SensorMessage{})
}
