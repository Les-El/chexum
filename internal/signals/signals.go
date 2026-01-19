// Package signals provides signal handling for graceful interruption.
//
// It catches SIGINT (Ctrl-C) and provides graceful shutdown:
// - First Ctrl-C: Stop processing, show status, run cleanup
// - Second Ctrl-C: Skip cleanup, exit immediately
package signals

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Handler manages signal handling for graceful interruption.
type Handler struct {
	interrupted    bool
	interruptCount int
	cleanup        func()
	cleanupTimeout time.Duration
	mu             sync.Mutex
	done           chan struct{}
	exitFunc       func(int) // Allow overriding os.Exit for testing
}

// NewSignalHandler creates a new signal handler with the given cleanup function.
func NewSignalHandler(cleanup func()) *Handler {
	return &Handler{
		cleanup:        cleanup,
		cleanupTimeout: 5 * time.Second,
		done:           make(chan struct{}),
		exitFunc:       os.Exit, // Default to os.Exit
	}
}

// SetCleanupTimeout sets the maximum time to wait for cleanup.
func (h *Handler) SetCleanupTimeout(d time.Duration) {
	h.cleanupTimeout = d
}

// Start begins listening for interrupt signals.
func (h *Handler) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigChan:
				h.handleInterrupt()
			case <-h.done:
				return
			}
		}
	}()
}

// Stop stops listening for signals.
func (h *Handler) Stop() {
	// Check if already stopped
	select {
	case <-h.done:
		// Already stopped
		return
	default:
		close(h.done)
		signal.Reset(os.Interrupt, syscall.SIGTERM)
	}
}

// handleInterrupt handles an interrupt signal.
func (h *Handler) handleInterrupt() {
	h.mu.Lock()
	h.interruptCount++
	count := h.interruptCount
	h.interrupted = true
	h.mu.Unlock()

	if count == 1 {
		// First interrupt: graceful shutdown
		fmt.Fprintln(os.Stderr, "\n\nInterrupted. Cleaning up...")

		if h.cleanup != nil {
			// Run cleanup with timeout
			done := make(chan struct{})
			go func() {
				h.cleanup()
				close(done)
			}()

			select {
			case <-done:
				fmt.Fprintln(os.Stderr, "Cleanup complete.")
			case <-time.After(h.cleanupTimeout):
				fmt.Fprintln(os.Stderr, "Cleanup timed out.")
			}
		}

		h.exitFunc(130) // 128 + SIGINT
	} else {
		// Second interrupt: immediate exit
		fmt.Fprintln(os.Stderr, "\nForced exit. Cleanup skipped.")
		h.exitFunc(130)
	}
}

// IsInterrupted returns whether an interrupt has been received.
func (h *Handler) IsInterrupted() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.interrupted
}

// InterruptCount returns the number of interrupts received.
func (h *Handler) InterruptCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.interruptCount
}

// Reset resets the interrupt state.
func (h *Handler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.interrupted = false
	h.interruptCount = 0
}
