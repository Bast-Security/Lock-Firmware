package main

import (
	"github.com/stianeikeland/go-rpio"
	"time"
)

func longBuzz(buzzerPin rpio.Pin) {
	buzzerPin.High()
	time.Sleep(time.Second)
	buzzerPin.Low()
}

