//go:build windows

package keyboard

import (
	"os"
	"testing"
	"time"
)

func TestTouchInjectionSmoke(t *testing.T) {
	if os.Getenv("QAA_TOUCH_SMOKE") != "1" {
		t.Skip("set QAA_TOUCH_SMOKE=1 to inject a short touch drag at the current cursor position")
	}

	if err := StartTouchScroll(); err != nil {
		t.Fatalf("start touch scroll: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := MoveTouchScroll(120); err != nil {
		t.Fatalf("move touch scroll: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := EndTouchScroll(); err != nil {
		t.Fatalf("end touch scroll: %v", err)
	}
}
