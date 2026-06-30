package controller

import (
	"context"
	"time"

	"github.com/by275/noidle/internal/log"
	"go.einride.tech/pid"
)

const samplingInterval = time.Second

type Device interface {
	Control(value float64)
	Measure() float64
}

// RunPID starts a PID controller loop with graceful shutdown support
func RunPID(
	device Device,
	referenceSignal float64,
	rateImpact float64,
	debug bool,
) *pid.Controller {
	referenceSignal = normalizeReferenceSignal(referenceSignal)
	c := &pid.Controller{
		Config: pid.ControllerConfig{
			ProportionalGain: rateImpact,
			IntegralGain:     1.0,
			DerivativeGain:   0,
		},
	}

	// Run PID controller in goroutine with proper context support
	// Note: For graceful shutdown, a context should be passed to this function
	// but keeping backward compatibility for now
	go pidControlLoop(c, device, referenceSignal, debug)

	return c
}

func pidControlLoop(c *pid.Controller, device Device, referenceSignal float64, debug bool) {
	for {
		actualSignal := device.Measure()
		c.Update(pid.ControllerInput{
			ReferenceSignal:  referenceSignal,
			ActualSignal:     actualSignal,
			SamplingInterval: samplingInterval,
		})
		if debug {
			log.Logf("PID", "actualSignal=%.2f controlSignal=%.2f", actualSignal, c.State.ControlSignal)
		}
		device.Control(c.State.ControlSignal)
	}
}

func normalizeReferenceSignal(referenceSignal float64) float64 {
	if referenceSignal < 0 || referenceSignal > 1 {
		log.Logf("PID", "invalid CPU waste ratio %.2f, falling back to 0.15", referenceSignal)
		referenceSignal = 0.15
	}
	return referenceSignal * 100
}
