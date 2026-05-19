package vibe

import (
	"irrlicht/core/adapters/inbound/agents/processlifecycle"
)

// DiscoverPID finds the Vibe process owning a session by matching "vibe" processes
// whose CWD equals the session's working directory. Since Vibe is a Python CLI
// tool, the process name is "vibe" (from the console_script entry point), so
// pgrep -x "vibe" works reliably.
//
// Vibe keeps transcript files open during the session lifetime, so we prefer
// discovering by transcript writer. However, Vibe sessions are stored in a
// central location (~/.vibe/logs/session/) not in the CWD, so we use CWD-based
// matching as the primary strategy since the transcript path won't be in the
// process's CWD.
func DiscoverPID(cwd, transcriptPath string, disambiguate func([]int) int) (int, error) {
	// Vibe stores sessions centrally, not in CWD, so transcript-based discovery
	// won't work (the transcript isn't in the process's CWD). Use CWD matching.
	return processlifecycle.DiscoverPIDByCWD(ProcessName, cwd, disambiguate)
}
