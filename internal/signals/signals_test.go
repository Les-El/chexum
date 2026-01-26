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

	if err := quick.Check(verifyRecoverableStateProperty, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func verifyRecoverableStateProperty(cleanupDuration uint8) bool {
	cleanupTime := time.Duration(cleanupDuration) * time.Millisecond
	tracker := &cleanupTracker{}

	handler := NewSignalHandler(tracker.runWithDuration(cleanupTime))
	handler.SetCleanupTimeout(500 * time.Millisecond)

	exitCalled := false
	handler.exitFunc = func(code int) { exitCalled = true }

	handler.Start()
	defer handler.Stop()
	handler.handleInterrupt()

	if !tracker.didStart() || !handler.IsInterrupted() || !exitCalled {
		return false
	}

	handler.Reset()
	if handler.IsInterrupted() {
		return false
	}

	if cleanupTime < 500*time.Millisecond && !tracker.didFinish() {
		return false
	}
	return true
}

func (c *cleanupTracker) runWithDuration(d time.Duration) func() {
	return func() {
		c.mu.Lock()
		c.started = true
		c.mu.Unlock()
		time.Sleep(d)
		c.mu.Lock()
		c.finished = true
		c.mu.Unlock()
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

func TestCleanupTimeout(t *testing.T) {
	tracker := &cleanupTracker{}
	handler := NewSignalHandler(tracker.run)
	handler.SetCleanupTimeout(50 * time.Millisecond)
	handler.exitFunc = func(code int) {}
	handler.Start()
	defer handler.Stop()

	start := time.Now()
	handler.handleInterrupt()
	duration := time.Since(start)

	if !tracker.didStart() {
		t.Error("Cleanup should have started")
	}
	if duration > 150*time.Millisecond {
		t.Errorf("Cleanup should have timed out, took %v", duration)
	}

	time.Sleep(200 * time.Millisecond)
	if !tracker.didFinish() {
		t.Error("Cleanup should have finished eventually")
	}
}

type cleanupTracker struct {
	started  bool
	finished bool
	mu       sync.Mutex
}

func (c *cleanupTracker) run() {
	c.mu.Lock()
	c.started = true
	c.mu.Unlock()
	time.Sleep(200 * time.Millisecond)
	c.mu.Lock()
	c.finished = true
	c.mu.Unlock()
}

func (c *cleanupTracker) didStart() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}

func (c *cleanupTracker) didFinish() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.finished
}

// TestDoubleCtrlC tests that second Ctrl-C skips cleanup.
func TestDoubleCtrlC(t *testing.T) {
	t.Run("First Interrupt runs cleanup", testFirstInterrupt)
	t.Run("Second Interrupt exits immediately", testSecondInterrupt)
}

func testFirstInterrupt(t *testing.T) {
	cleanupCount := 0
	exitCount := 0
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupCount++
		mu.Unlock()
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {
		mu.Lock()
		exitCount++
		mu.Unlock()
	}
	handler.Start()
	defer handler.Stop()

	handler.handleInterrupt()

	if handler.InterruptCount() != 1 || cleanupCount != 1 || exitCount != 1 {
		t.Errorf("Unexpected state after first interrupt: count=%d, cleanup=%d, exit=%d",
			handler.InterruptCount(), cleanupCount, exitCount)
	}
}

func testSecondInterrupt(t *testing.T) {
	cleanupCount := 0
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupCount++
		mu.Unlock()
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {}
	handler.Start()
	defer handler.Stop()

	// Simulate already interrupted
	handler.mu.Lock()
	handler.interruptCount = 1
	handler.interrupted = true
	handler.mu.Unlock()

	handler.handleInterrupt()

	if handler.InterruptCount() != 2 || cleanupCount != 0 {
		t.Errorf("Unexpected state after second interrupt: count=%d, cleanup=%d",
			handler.InterruptCount(), cleanupCount)
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
	filesProcessed, cleanupRan := runIntegrationScenario()

	if filesProcessed != 5 {
		t.Errorf("Expected 5 files processed, got %d", filesProcessed)
	}
	if !cleanupRan {
		t.Error("Cleanup should have run")
	}
}

func runIntegrationScenario() (int, bool) {
	filesProcessed := 0
	cleanupRan := false
	var mu sync.Mutex

	cleanup := func() {
		mu.Lock()
		cleanupRan = true
		mu.Unlock()
	}

	handler := NewSignalHandler(cleanup)
	handler.exitFunc = func(code int) {}
	handler.Start()
	defer handler.Stop()

	for i := 0; i < 10; i++ {
		if handler.IsInterrupted() {
			break
		}
		mu.Lock()
		filesProcessed++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		if i == 4 {
			go handler.handleInterrupt()
			time.Sleep(20 * time.Millisecond)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	return filesProcessed, cleanupRan
}

func TestNewSignalHandler(t *testing.T) {
	h := NewSignalHandler(nil)
	if h == nil {
		t.Fatal("NewSignalHandler returned nil")
	}
}

func TestStart(t *testing.T) {
	h := NewSignalHandler(nil)
	h.exitFunc = func(int) {}
	h.Start()
	h.Stop()
}

func TestStop(t *testing.T) {
	TestStopSignalHandler(t)
}

func TestInterruptCount(t *testing.T) {
	h := NewSignalHandler(nil)
	h.exitFunc = func(int) {}
	if h.InterruptCount() != 0 {
		t.Error("Expected 0")
	}
	h.handleInterrupt()
	if h.InterruptCount() != 1 {
		t.Error("Expected 1")
	}
}
