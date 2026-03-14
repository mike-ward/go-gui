//go:build !linux

// Package atspi provides AT-SPI2 accessibility on Linux.
// This file is a no-op stub for non-Linux platforms.
package atspi

import "github.com/mike-ward/go-gui/gui"

// Bridge is a no-op on non-Linux platforms.
type Bridge struct{}

func (b *Bridge) Init(_ func(action, index int))               {}
func (b *Bridge) Sync(_ []gui.A11yNode, _, _ int)              {}
func (b *Bridge) Destroy()                                     {}
func (b *Bridge) Announce(_ string)                            {}
