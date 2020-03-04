package main

import (
	"github.com/stianeikeland/go-rpio"
	"time"
)

func buzz(buzzerPin rpio.Pin, dur time.Duration) {
	start := time.Now()
	buzzerPin.High()
	for time.Now().Sub(start) < dur { }
	buzzerPin.Low()
}

