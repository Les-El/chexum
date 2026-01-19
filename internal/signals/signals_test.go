package signals

import (
	"sync"
	"testing"
	"testing/quick"
	"time"
)

// TestPropertyInterruptedOperationsLeaveRecoverableState tests Property 14:
// For any operation interrupted by SIGINT, the system should be in a state
// where the operation can be safely retried.
//
// Feature: cli-guidelines-review, Property 14: Interrupted operations leave recoverable state
// Validates: Requirements 12.4, 12.5
func TestPropertyInterruptedOperationsLeaveRecoverableState(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// Property: After interruption, cleanup should run (or timeout gracefully),
	// and the handler should be in a consistent state that allows retry
	property := func(cleanupDuration uint8) bool {
		// Convert to milliseconds (0-255ms)
		cleanupTime := time.Duration(cleanupDuration) * time.Millisecond

		// Track cleanup state
		cleanupRan := false
		cleanupComplete := false
		var mu sync.Mutex

		cleanup := func() {
			mu.Lock()
			cleanupRan = true
			mu.Unlock()

			// Simulate cleanup work
			time.Sleep(cleanupTime)

			mu.Lock()
			cleanupComplete = true
			mu.Unlock()
		}

		handler := NewSignalHandler(cleanup)
		handler.SetCleanupTimeout(500 * time.Millisecond)
		
		// Override exit function for testing
		exitCalled := false
		handler.exitFunc = func(code int) {
			exitCalled = true
		}

		// Start the handler
		handler.Start()
		defer handler.Stop()

		// Simulate interrupt
		handler.handleInterrupt()

		// Check state after interrupt
		mu.Lock()
		ranCleanup := cleanupRan
		completedCleanup := cleanupComplete
		mu.Unlock()

		// Verify recoverable state:
		// 1. Cleanup was attempted
		// 2. Handler is marked as interrupted
		// 3. Handler can be reset for retry
		// 4. Exit was called
		if !ranCleanup {
			return false
		}

		if !handler.IsInterrupted() {
			return false
		}
		
		if !exitCalled {
			return false
		}

		// Reset should allow retry
		handler.Reset()
		if handler.IsInterrupted() {
			return false
		}

		// If cleanup was fast enough, it should have completed
		if cleanupTime < 500*time.Millisecond && !completedCleanup {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestSIGINTHandling tests that SIGINT is caught and handled correctly.
func TestSIGINTHandling(t *testing.T) {
	cleanupRan := false
	cleanup := func() {
		cleanupRan = true
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Initially not interrupted
	if handler.IsInterrupted() {
		t.Error("Handler should not be interrupted initially")
	}

	// Simulate first interrupt
	handler.handleInterrupt()

	// Should be marked as interrupted
	if !handler.IsInterrupted() {
		t.Error("Handler should be interrupted after first signal")
	}

	// Cleanup should have run
	if !cleanupRan {
		t.Error("Cleanup should have run after first interrupt")
	}

	// Interrupt count should be 1
	if handler.InterruptCount() != 1 {
		t.Errorf("Expected interrupt count 1, got %d", handler.InterruptCount())
	}
}

// TestCleanupTimeout tests that cleanup respects the timeout.
func TestCleanupTimeout(t *testing.T) {
	cleanupStarted := false
	cleanupFinished := false
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupStarted = true
		mu.Unlock()

		// Simulate long cleanup (longer than timeout)
		time.Sleep(200 * time.Millisecond)

		mu.Lock()
		cleanupFinished = true
		mu.Unlock()
	}

	handler := NewSignalHandler(cleanup)
	handler.SetCleanupTimeout(50 * time.Millisecond)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Simulate interrupt
	start := time.Now()
	handler.handleInterrupt()
	duration := time.Since(start)

	// Check that cleanup started
	mu.Lock()
	started := cleanupStarted
	mu.Unlock()

	if !started {
		t.Error("Cleanup should have started")
	}

	// Should have timed out (not waited for full cleanup)
	// Allow some margin for timing
	if duration > 150*time.Millisecond {
		t.Errorf("Cleanup should have timed out, but took %v", duration)
	}

	// Give cleanup goroutine time to finish
	time.Sleep(200 * time.Millisecond)

	// Cleanup should eventually finish (in background)
	mu.Lock()
	finished := cleanupFinished
	mu.Unlock()

	if !finished {
		t.Error("Cleanup should have finished in background")
	}
}

// TestDoubleCtrlC tests that second Ctrl-C skips cleanup.
func TestDoubleCtrlC(t *testing.T) {
	cleanupCount := 0
	exitCount := 0
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupCount++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {
		mu.Lock()
		exitCount++
		mu.Unlock()
	}
	handler.Start()
	defer handler.Stop()

	// First interrupt - should run cleanup
	handler.handleInterrupt()

	mu.Lock()
	count1 := handler.InterruptCount()
	cleanup1 := cleanupCount
	exit1 := exitCount
	mu.Unlock()

	if count1 != 1 {
		t.Errorf("Expected interrupt count 1 after first signal, got %d", count1)
	}

	if cleanup1 != 1 {
		t.Errorf("Expected cleanup to run once, got %d", cleanup1)
	}

	if exit1 != 1 {
		t.Errorf("Expected exit to be called once, got %d", exit1)
	}

	// Create a new handler for second interrupt test
	handler2 := NewSignalHandler(cleanup)
	handler2.exitFunc = func(code int) {
		mu.Lock()
		exitCount++
		mu.Unlock()
	}
	handler2.Start()
	defer handler2.Stop()

	// Manually set interrupt count to 2 to simulate second interrupt
	handler2.mu.Lock()
	handler2.interruptCount = 1
	handler2.interrupted = true
	handler2.mu.Unlock()

	// Second interrupt should exit immediately without running cleanup again
	handler2.handleInterrupt()

	mu.Lock()
	cleanup2 := cleanupCount
	count2 := handler2.InterruptCount()
	mu.Unlock()

	// Cleanup should still be 1 (not run for second interrupt)
	if cleanup2 != 1 {
		t.Errorf("Expected cleanup to run only once total, got %d", cleanup2)
	}

	// Interrupt count should be 2
	if count2 != 2 {
		t.Errorf("Expected interrupt count 2 after second signal, got %d", count2)
	}
}

// TestIsInterrupted tests the IsInterrupted method.
func TestIsInterrupted(t *testing.T) {
	handler := NewSignalHandler(nil)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	if handler.IsInterrupted() {
		t.Error("Should not be interrupted initially")
	}

	handler.handleInterrupt()

	if !handler.IsInterrupted() {
		t.Error("Should be interrupted after signal")
	}
}

// TestReset tests that Reset clears the interrupt state.
func TestReset(t *testing.T) {
	handler := NewSignalHandler(nil)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Trigger interrupt
	handler.handleInterrupt()

	if !handler.IsInterrupted() {
		t.Error("Should be interrupted")
	}

	if handler.InterruptCount() != 1 {
		t.Errorf("Expected interrupt count 1, got %d", handler.InterruptCount())
	}

	// Reset
	handler.Reset()

	if handler.IsInterrupted() {
		t.Error("Should not be interrupted after reset")
	}

	if handler.InterruptCount() != 0 {
		t.Errorf("Expected interrupt count 0 after reset, got %d", handler.InterruptCount())
	}
}

// TestNilCleanup tests that handler works with nil cleanup function.
func TestNilCleanup(t *testing.T) {
	handler := NewSignalHandler(nil)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Should not panic with nil cleanup
	handler.handleInterrupt()

	if !handler.IsInterrupted() {
		t.Error("Should be interrupted even with nil cleanup")
	}
}

// TestConcurrentInterrupts tests thread safety of interrupt handling.
func TestConcurrentInterrupts(t *testing.T) {
	handler := NewSignalHandler(func() {
		time.Sleep(10 * time.Millisecond)
	})
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Try to trigger multiple interrupts concurrently
	// Only the first should run cleanup
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Note: In real usage, only one would succeed before os.Exit
			// Here we're testing the state management
			if handler.InterruptCount() == 0 {
				handler.handleInterrupt()
			}
		}()
	}

	wg.Wait()

	// Should be interrupted
	if !handler.IsInterrupted() {
		t.Error("Should be interrupted")
	}

	// Count should be at least 1
	if handler.InterruptCount() < 1 {
		t.Error("Should have at least one interrupt")
	}
}

// TestSetCleanupTimeout tests setting custom cleanup timeout.
func TestSetCleanupTimeout(t *testing.T) {
	handler := NewSignalHandler(func() {
		time.Sleep(100 * time.Millisecond)
	})

	// Set custom timeout
	customTimeout := 200 * time.Millisecond
	handler.SetCleanupTimeout(customTimeout)

	if handler.cleanupTimeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, handler.cleanupTimeout)
	}
}

// TestStopSignalHandler tests that Stop properly cleans up.
func TestStopSignalHandler(t *testing.T) {
	handler := NewSignalHandler(nil)
	handler.Start()

	// Stop should not panic
	handler.Stop()

	// Calling Stop again should not panic
	handler.Stop()
}

// TestSignalHandlerIntegration tests a realistic usage scenario.
func TestSignalHandlerIntegration(t *testing.T) {
	// Simulate a realistic scenario with file processing
	filesProcessed := 0
	cleanupRan := false
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupRan = true
		mu.Unlock()
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {} // Mock exit
	handler.Start()
	defer handler.Stop()

	// Simulate processing files
	for i := 0; i < 10; i++ {
		if handler.IsInterrupted() {
			break
		}

		mu.Lock()
		filesProcessed++
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		// Simulate interrupt after processing 5 files
		if i == 4 {
			go handler.handleInterrupt()
			time.Sleep(20 * time.Millisecond) // Let interrupt handler run
		}
	}

	mu.Lock()
	processed := filesProcessed
	cleaned := cleanupRan
	mu.Unlock()

	// Should have processed 5 files before interrupt
	if processed != 5 {
		t.Errorf("Expected 5 files processed, got %d", processed)
	}

	// Cleanup should have run
	if !cleaned {
		t.Error("Cleanup should have run")
	}

	// Should be interrupted
	if !handler.IsInterrupted() {
		t.Error("Should be interrupted")
	}
}
