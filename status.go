package main

import (
	"time"
)

// Status holds the current status of the tool
type Status struct {
	LastCheck   time.Time
	LastBalance float64
	Status      string
	Message     string
}
