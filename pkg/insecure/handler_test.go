package insecure

import (
	"testing"
)

func TestNewInsecureHandler(t *testing.T) {
	handler := NewInsecureHandler()

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}

	if handler.enabled {
		t.Error("Expected handler to be initially disabled")
	}

	if len(handler.registries) != 0 {
		t.Error("Expected empty registries map")
	}

	if len(handler.warnOnce) != 0 {
		t.Error("Expected empty warnOnce map")
	}
}

func TestInsecureHandler_GlobalMode(t *testing.T) {
	handler := NewInsecureHandler()

	// Initially disabled
	if handler.IsInsecure("any.registry") {
		t.Error("Expected IsInsecure to return false initially")
	}

	if handler.IsGlobalInsecure() {
		t.Error("Expected IsGlobalInsecure to return false initially")
	}

	// Enable global insecure mode
	handler.Enable()

	// Should be insecure for all registries
	if !handler.IsInsecure("registry1.example.com") {
		t.Error("Expected IsInsecure to return true for registry1 in global mode")
	}

	if !handler.IsInsecure("registry2.example.com") {
		t.Error("Expected IsInsecure to return true for registry2 in global mode")
	}

	if !handler.IsGlobalInsecure() {
		t.Error("Expected IsGlobalInsecure to return true")
	}

	// Check registries list
	registries := handler.GetInsecureRegistries()
	if len(registries) != 1 || registries[0] != "*" {
		t.Errorf("Expected registries list to be ['*'], got %v", registries)
	}
}

func TestInsecureHandler_RegistrySpecific(t *testing.T) {
	handler := NewInsecureHandler()

	// Enable for specific registries
	handler.Enable("registry1.example.com", "registry2.example.com")

	// Should be insecure only for specified registries
	if !handler.IsInsecure("registry1.example.com") {
		t.Error("Expected IsInsecure to return true for registry1")
	}

	if !handler.IsInsecure("registry2.example.com") {
		t.Error("Expected IsInsecure to return true for registry2")
	}

	if handler.IsInsecure("registry3.example.com") {
		t.Error("Expected IsInsecure to return false for registry3")
	}

	if handler.IsGlobalInsecure() {
		t.Error("Expected IsGlobalInsecure to return false for registry-specific mode")
	}

	// Check registries list
	registries := handler.GetInsecureRegistries()
	if len(registries) != 2 {
		t.Errorf("Expected 2 registries, got %d", len(registries))
	}

	// Check that both registries are in the list (order may vary)
	hasRegistry1 := false
	hasRegistry2 := false
	for _, reg := range registries {
		if reg == "registry1.example.com" {
			hasRegistry1 = true
		} else if reg == "registry2.example.com" {
			hasRegistry2 = true
		}
	}

	if !hasRegistry1 {
		t.Error("Expected registry1.example.com in registries list")
	}

	if !hasRegistry2 {
		t.Error("Expected registry2.example.com in registries list")
	}
}

func TestInsecureHandler_Disable(t *testing.T) {
	handler := NewInsecureHandler()

	// Enable global mode
	handler.Enable()
	if !handler.IsInsecure("test.registry") {
		t.Error("Expected insecure mode to be enabled")
	}

	// Disable
	handler.Disable()
	if handler.IsInsecure("test.registry") {
		t.Error("Expected insecure mode to be disabled")
	}

	if handler.IsGlobalInsecure() {
		t.Error("Expected global insecure mode to be disabled")
	}

	registries := handler.GetInsecureRegistries()
	if registries != nil {
		t.Error("Expected GetInsecureRegistries to return nil when disabled")
	}
}

func TestInsecureHandler_WarnOnce(t *testing.T) {
	handler := NewInsecureHandler()

	// First warning should set the flag
	handler.WarnRegistry("test.registry")
	if !handler.warnOnce["test.registry"] {
		t.Error("Expected warnOnce flag to be set for test.registry")
	}

	// Subsequent calls should not change state
	handler.WarnRegistry("test.registry")
	if !handler.warnOnce["test.registry"] {
		t.Error("Expected warnOnce flag to remain set")
	}

	// Different registry should get its own flag
	handler.WarnRegistry("other.registry")
	if !handler.warnOnce["other.registry"] {
		t.Error("Expected warnOnce flag to be set for other.registry")
	}

	// Original registry flag should still be set
	if !handler.warnOnce["test.registry"] {
		t.Error("Expected original warnOnce flag to remain set")
	}
}

func TestInsecureHandler_WarnGlobal(t *testing.T) {
	handler := NewInsecureHandler()

	// First warning should set the flag
	handler.WarnGlobal()
	if !handler.warnOnce["_global"] {
		t.Error("Expected warnOnce flag to be set for global warning")
	}

	// Subsequent calls should not change state
	handler.WarnGlobal()
	if !handler.warnOnce["_global"] {
		t.Error("Expected global warnOnce flag to remain set")
	}
}

func TestInsecureHandler_StateTransitions(t *testing.T) {
	handler := NewInsecureHandler()

	// Start disabled
	if handler.enabled {
		t.Error("Expected handler to start disabled")
	}

	// Enable globally
	handler.Enable()
	if !handler.enabled {
		t.Error("Expected handler to be enabled after Enable()")
	}
	if !handler.IsGlobalInsecure() {
		t.Error("Expected global insecure mode")
	}

	// Switch to registry-specific
	handler.Enable("specific.registry")
	if !handler.enabled {
		t.Error("Expected handler to remain enabled")
	}
	if handler.IsGlobalInsecure() {
		t.Error("Expected registry-specific mode, not global")
	}
	if !handler.IsInsecure("specific.registry") {
		t.Error("Expected specific.registry to be insecure")
	}

	// Disable completely
	handler.Disable()
	if handler.enabled {
		t.Error("Expected handler to be disabled")
	}
	if handler.IsInsecure("specific.registry") {
		t.Error("Expected no registries to be insecure after disable")
	}
}

func TestInsecureHandler_MultipleRegistries(t *testing.T) {
	handler := NewInsecureHandler()

	registries := []string{
		"registry1.example.com",
		"registry2.example.com",
		"localhost:5000",
		"my-registry.local",
	}

	// Enable for multiple registries
	handler.Enable(registries...)

	// All specified registries should be insecure
	for _, reg := range registries {
		if !handler.IsInsecure(reg) {
			t.Errorf("Expected registry %s to be insecure", reg)
		}
	}

	// Non-specified registry should not be insecure
	if handler.IsInsecure("other.registry") {
		t.Error("Expected non-specified registry to not be insecure")
	}

	// Should not be global mode
	if handler.IsGlobalInsecure() {
		t.Error("Expected registry-specific mode, not global")
	}

	// Get registries list should match
	insecureRegistries := handler.GetInsecureRegistries()
	if len(insecureRegistries) != len(registries) {
		t.Errorf("Expected %d registries, got %d", len(registries), len(insecureRegistries))
	}
}