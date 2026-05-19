// Package vibe provides an inbound adapter that monitors Mistral Vibe coding
// agent session transcripts under ~/.vibe/logs/session/ (or $VIBE_HOME/logs/session/).
//
// Mistral Vibe (https://github.com/mistralai/mistral-vibe) is a CLI coding agent
// that stores session transcripts as JSONL files in session_{timestamp}_{id}/
// directories, with each session containing a messages.jsonl file.
package vibe

import (
	"log"
	"os"
	"path/filepath"
)

// AdapterName identifies sessions originating from Mistral Vibe.
const AdapterName = "vibe"

// ProcessName is the OS-level executable name for Mistral Vibe, used by
// the process lifecycle scanner to detect running instances via pgrep -x.
// Vibe is a Python CLI tool installed via pip, so the process name is "vibe".
const ProcessName = "vibe"

// defaultRootDir is the path relative to $HOME where Vibe stores session
// transcripts by default. Sessions live under logs/session/
// session_{timestamp}_{short_session_id}/messages.jsonl.
const defaultRootDir = ".vibe/logs/session"

// sessionDirEnvVar is the upstream Vibe env var that relocates the session-
// transcript root. When set, it should point to the parent of logs/session.
// We construct the full path as $VIBE_HOME/logs/session.
const sessionDirEnvVar = "VIBE_HOME"

// sessionsDir returns the directory the Vibe adapter should watch.
// Non-absolute env values are rejected so a misconfigured path surfaces
// in logs instead of silently watching the wrong place.
func sessionsDir() string {
	if v := os.Getenv(sessionDirEnvVar); v != "" {
		cleaned := filepath.Clean(v)
		if filepath.IsAbs(cleaned) {
			// VIBE_HOME points to the .vibe directory, so we append logs/session
			return filepath.Join(cleaned, "logs", "session")
		}
		log.Printf("vibe: ignoring %s=%q — must be an absolute path (no shell expansion)", sessionDirEnvVar, v)
	}
	return defaultRootDir
}
