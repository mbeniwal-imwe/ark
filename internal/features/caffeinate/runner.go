package caffeinate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Mode string

const (
	ModeWiggle     Mode = "wiggle"     // use osascript to nudge cursor/keypress
	ModeCaffeinate      = "caffeinate" // fallback to macOS caffeinate tool
)

type Runner struct {
	ConfigDir string
	Interval  time.Duration
	Mode      Mode
}

func (r *Runner) pidFile() string {
	return filepath.Join(r.ConfigDir, "data", "caffeinate.pid")
}

func (r *Runner) isRunning() (bool, int, error) {
	pidBytes, err := os.ReadFile(r.pidFile())
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}
	pidStr := strings.TrimSpace(string(pidBytes))
	if pidStr == "" {
		return false, 0, nil
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false, 0, nil
	}

	// Use ps command to check if process exists (more reliable on macOS)
	cmd := exec.Command("ps", "-p", pidStr, "-o", "pid=")
	output, err := cmd.Output()
	if err != nil {
		return false, 0, nil
	}

	// Check if ps returned a valid PID
	if strings.TrimSpace(string(output)) == pidStr {
		return true, pid, nil
	}
	return false, 0, nil
}

func (r *Runner) writePID(pid int) error {
	if err := os.MkdirAll(filepath.Dir(r.pidFile()), 0700); err != nil {
		return err
	}
	return os.WriteFile(r.pidFile(), []byte(strconv.Itoa(pid)), 0600)
}

func (r *Runner) clearPID() { _ = os.Remove(r.pidFile()) }

// Start launches a background process to keep the device awake
func (r *Runner) Start() error {
	running, _, err := r.isRunning()
	if err != nil {
		return err
	}
	if running {
		return errors.New("caffeinate already running")
	}

	// Re-exec self with internal flag to run the loop
	self, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{"caffeinate", "_run", "--interval", fmt.Sprintf("%d", int(r.Interval.Seconds())), "--mode", string(r.Mode)}
	cmd := exec.Command(self, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := r.writePID(cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	return nil
}

func (r *Runner) Stop() error {
	running, pid, err := r.isRunning()
	if err != nil {
		return err
	}
	if !running {
		return errors.New("caffeinate not running")
	}
	// Attempt graceful kill
	_ = exec.Command("kill", strconv.Itoa(pid)).Run()
	r.clearPID()
	return nil
}

func (r *Runner) Status() (string, error) {
	running, pid, err := r.isRunning()
	if err != nil {
		return "", err
	}
	if running {
		return fmt.Sprintf("running (pid %d)", pid), nil
	}
	return "stopped", nil
}

// RunLoop is invoked by the re-exec path: ark caffeinate _run --interval N --mode M
func RunLoop(intervalSec int, mode Mode) error {
	interval := time.Duration(intervalSec) * time.Second
	if interval < 5*time.Second {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if mode == ModeWiggle {
			// Try a harmless keypress (shift down/up) using osascript (avoids mouse permissions for many setups)
			// If that fails, try moving the cursor by 1px and back (also via osascript)
			if err := exec.Command("osascript", "-e", `tell application "System Events" to key down shift`).Run(); err == nil {
				_ = exec.Command("osascript", "-e", `tell application "System Events" to key up shift`).Run()
			} else {
				_ = exec.Command("osascript", "-e", `do shell script "python3 - <<'PY'\nimport Quartz, time\nloc = Quartz.CGEventGetLocation(Quartz.CGEventCreate(None))\nQuartz.CGWarpMouseCursorPosition((loc.x+1, loc.y))\nQuartz.CGWarpMouseCursorPosition((loc.x, loc.y))\nPY"`).Run()
			}
		} else {
			// Fallback to macOS caffeinate for the interval window
			_ = exec.Command("caffeinate", "-u", "-t", fmt.Sprintf("%d", int(interval.Seconds()))).Run()
		}
		<-ticker.C
	}
}
