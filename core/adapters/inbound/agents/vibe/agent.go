package vibe

import (
	"irrlicht/core/domain/agent"
)

// Mistral Vibe icon — purple diamond shape representing the Mistral brand.
// Uses a purple color (#8B5CF6) that reads well in both light and dark themes.
const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 100 100">
  <path d="M50 10 L90 50 L50 90 L10 50 Z" fill="#8B5CF6"/>
</svg>`

// Agent returns the Mistral Vibe adapter registration.
func Agent() agent.Agent {
	return agent.Agent{
		Identity: agent.Identity{
			Name:         AdapterName,
			DisplayName:  "Mistral Vibe",
			IconSVGLight: iconSVG,
			IconSVGDark:  iconSVG,
		},
		Process: agent.Process{
			Match:         agent.ExactName{Name: ProcessName},
			PIDForSession: DiscoverPID,
		},
		Source: agent.FilesUnderRoot{
			Dir: sessionsDir(),
			Parser: agent.JSONLineParser{
				NewParser: func() agent.LineParser { return &Parser{} },
			},
		},
	}
}
