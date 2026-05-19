package vibe

import (
	"testing"
)

func TestAgent(t *testing.T) {
	agent := Agent()

	if agent.Identity.Name != AdapterName {
		t.Errorf("Identity.Name = %q, want %q", agent.Identity.Name, AdapterName)
	}

	if agent.Identity.DisplayName != "Mistral Vibe" {
		t.Errorf("Identity.DisplayName = %q, want %q", agent.Identity.DisplayName, "Mistral Vibe")
	}

	if agent.Identity.IconSVGLight == "" {
		t.Error("IconSVGLight is empty")
	}

	if agent.Identity.IconSVGDark == "" {
		t.Error("IconSVGDark is empty")
	}

	// Check PID discovery function is set
	if agent.Process.PIDForSession == nil {
		t.Error("PIDForSession is nil")
	}

	// Check source configuration is set
	if agent.Source == nil {
		t.Fatal("Source is nil")
	}
}
