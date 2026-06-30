//go:build linux

package controller

import (
	"os"
	"syscall"

	"github.com/by275/noidle/internal/log"
)

const ioprioClassShift = 13

type ioprioClass = int

const (
	ioprioClassRT ioprioClass = (iota + 1) << ioprioClassShift
	ioprioClassBE
	ioprioClassIdle
)

const (
	ioprioWhoProcess = iota + 1
	ioprioWhoPGRP
	ioprioWhoUser
)

func ioprioSet(class ioprioClass, value int) error {
	// Properly handle syscall errors instead of ignoring them
	_, _, errno := syscall.Syscall(syscall.SYS_IOPRIO_SET,
		uintptr(ioprioWhoPGRP), uintptr(os.Getpid()),
		uintptr(class)|uintptr(value))

	if errno != 0 {
		log.Logf("IOPRIO", "Failed to set I/O priority: %v", errno)
		return errno
	}
	return nil
}

func SetPriority(nice int) error {
	// https://github.com/syncthing/syncthing/issues/4628
	// Move ourselves to a new process group so that we can use the process
	// group variants of Setpriority etc. to affect all of our threads in one
	// go. If this fails, bail, so that we don't affect things we shouldn't.
	if err := syscall.Setpgid(0, 0); err != nil {
		log.Logf("PRIO", "Failed to set process group: %v", err)
		return err
	}

	err := syscall.Setpriority(syscall.PRIO_PGRP, 0, nice)
	if err != nil {
		log.Logf("PRIO", "Failed to set priority: %v", err)
	}
	return err
}

func SetWorstPriority() {
	if err := SetPriority(19); err != nil {
		log.Logf("PRIO", "Warning: failed to set CPU priority: %v", err)
	}

	// 0 through 7 being the range. Log errors instead of ignoring.
	if err := ioprioSet(ioprioClassIdle, 7); err != nil {
		log.Logf("PRIO", "Warning: failed to set I/O priority: %v", err)
	}
}
