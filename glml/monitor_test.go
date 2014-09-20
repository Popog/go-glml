// Copyright © 2012 Popog
package glml

import (
	"testing"
)

// Test if the default monitor has basic functionality
func TestMonitor_Default(t *testing.T) {
	monitor := GetDefaultMonitor()
	if !monitor.IsValid() {
		t.Error("default monitor is not valid")
	}
	if !monitor.IsDefault() {
		t.Error("default monitor is not default")
	}
	for _, mode := range monitor.GetFullscreenVideoModes() {
		if !monitor.SupportsMode(mode) {
			t.Error("default monitor does not support mode")
		}
	}

}

func TestMonitor_GetMonitors(t *testing.T) {
	var default_count int
	for _, monitor := range GetMonitors() {
		if monitor == nil {
			t.Error("monitor is nil")
		}
		if !monitor.IsValid() {
			t.Error("monitor is not valid")
		}
		if monitor.IsDefault() {
			default_count++
		}
	}

	if default_count != 1 {
		t.Errorf("incorrect number of default monitors (%d)", default_count)
	}
}

func TestMonitor_Modes(t *testing.T) {
	for _, monitor := range GetMonitors() {
		for _, mode := range monitor.GetFullscreenVideoModes() {
			if !monitor.SupportsMode(mode) {
				t.Error("monitor does not support mode")
			}
		}
	}
}
