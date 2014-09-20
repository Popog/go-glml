// Copyright © 2012 Popog
package glml

import (
	"testing"
	"time"
)

func TestWindowStyles(t *testing.T) {
	m := GetDefaultMonitor()
	thread := CreateThread()
	for _, mode := range m.GetFullscreenVideoModes() {
		window, err := CreateWindow(m, mode, "Test", WindowStyleFullscreen, ContextSettingsDefault)
		if err != nil {
			t.Error(err)
		}
		thread.SetActive(window)
	window_SetActive_Errors:
		for {
			select {
			case err := <-window.Errors():
				if err.Fatal() {
					t.Fatal(err)
				} else {
					t.Error(err)
				}
			default:
				break window_SetActive_Errors
			}
		}

		for i := 0; i < 600; i++ {
			finished := make(chan bool)
			window.Commands() <- func(thread *Thread, t Threadable) ThreadError {
				w := t.(*Window)

				// Poll the events
				if _, err := w.ThreadPollEvents(thread, false); len(err) != 0 {
					return err[len(err)-1] // return the last error, it's the most fatal one
				}

				if err := w.ThreadSwapBuffers(); err != nil {
					return err
				}

				finished <- true
				return nil
			}

			select {
			case err := <-window.Errors():
				if err.Fatal() {
					t.Fatal(err)
				} else {
					t.Error(err)
				}
			case <-finished:
			}
			time.Sleep(10 * time.Millisecond)
		}

		window.Close()

		for i := 0; i < 400; i++ {
			time.Sleep(10 * time.Millisecond)
		}
	}
	thread.Close()
}
