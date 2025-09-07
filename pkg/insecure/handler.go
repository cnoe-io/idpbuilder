package insecure

import (
	"fmt"
	"os"
	"strings"
)

// InsecureHandler manages the --insecure flag behavior
type InsecureHandler struct {
	enabled     bool
	registries  map[string]bool
	warnOnce    map[string]bool
}

// NewInsecureHandler creates a new insecure mode handler
func NewInsecureHandler() *InsecureHandler {
	return &InsecureHandler{
		enabled:    false,
		registries: make(map[string]bool),
		warnOnce:   make(map[string]bool),
	}
}

// Enable activates insecure mode
func (h *InsecureHandler) Enable(registries ...string) {
	h.enabled = true

	if len(registries) == 0 {
		// Global insecure mode
		h.WarnGlobal()
	} else {
		// Registry-specific insecure mode
		for _, reg := range registries {
			h.registries[reg] = true
			h.WarnRegistry(reg)
		}
	}
}

// Disable deactivates insecure mode
func (h *InsecureHandler) Disable() {
	h.enabled = false
	h.registries = make(map[string]bool)
	// Keep warnOnce to avoid repeated warnings if re-enabled
}

// IsInsecure checks if insecure mode is enabled for a registry
func (h *InsecureHandler) IsInsecure(registry string) bool {
	if !h.enabled {
		return false
	}

	if len(h.registries) == 0 {
		// Global insecure mode
		return true
	}
	return h.registries[registry]
}

// IsGlobalInsecure returns true if global insecure mode is enabled
func (h *InsecureHandler) IsGlobalInsecure() bool {
	return h.enabled && len(h.registries) == 0
}

// GetInsecureRegistries returns the list of registries with insecure mode enabled
func (h *InsecureHandler) GetInsecureRegistries() []string {
	if !h.enabled {
		return nil
	}

	if len(h.registries) == 0 {
		return []string{"*"} // Global mode indicator
	}

	var registries []string
	for reg := range h.registries {
		registries = append(registries, reg)
	}
	return registries
}

// WarnGlobal displays a warning for global insecure mode
func (h *InsecureHandler) WarnGlobal() {
	if !h.warnOnce["_global"] {
		fmt.Fprintln(os.Stderr, strings.Repeat("⚠", 10))
		fmt.Fprintln(os.Stderr, "WARNING: Running in INSECURE mode")
		fmt.Fprintln(os.Stderr, "Certificate validation is DISABLED for ALL registries")
		fmt.Fprintln(os.Stderr, "This should ONLY be used in development environments")
		fmt.Fprintln(os.Stderr, strings.Repeat("⚠", 10))
		h.warnOnce["_global"] = true
	}
}

// WarnRegistry displays a warning for registry-specific insecure mode
func (h *InsecureHandler) WarnRegistry(registry string) {
	if !h.warnOnce[registry] {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Certificate validation disabled for %s\n", registry)
		h.warnOnce[registry] = true
	}
}