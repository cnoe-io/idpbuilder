# Effort 1.1.1: Push Command Skeleton Implementation Plan

## 🚨 CRITICAL EFFORT METADATA (FROM WAVE PLAN)
**Branch**: `phase1/wave1/effort-1.1.1-push-command-skeleton`
**Can Parallelize**: Yes
**Parallel With**: [1.1.2, 1.1.3]
**Size Estimate**: ~350 lines (well under 800 limit)
**Dependencies**: None (foundational effort)
**Base Branch**: main

## Overview
- **Effort**: Create basic push command structure with cobra
- **Phase**: 1, Wave: 1
- **Estimated Size**: ~350 lines
- **Implementation Time**: 3-4 hours

## Scope Statement (R371 Compliance)
This effort implements ONLY the basic push command structure and registration. It does NOT include:
- Authentication logic (effort 1.1.2)
- TLS configuration (effort 1.1.3)
- Registry client implementation (Phase 2)
- Image handling logic (Phase 3)
- Actual push functionality (Phase 4)

## Theme Coherence (R372 Compliance)
**Single Theme**: Basic CLI command structure and registration
- All changes support creating the push command skeleton
- No mixed concerns with auth, TLS, or registry logic
- Pure CLI framework setup

## Architectural Compliance (R362)
**Framework Used**: spf13/cobra (existing in project)
- Following existing IDPBuilder command patterns
- No custom CLI framework implementation
- Using standard cobra command structure

## Library Version Requirements (R381)
**Locked Dependencies (DO NOT UPDATE)**:
- github.com/spf13/cobra: Use existing version in go.mod
- github.com/spf13/viper: Use existing version if needed

**New Dependencies**: None for this effort

## File Structure
```
idpbuilder-gitea-push/
├── cmd/
│   ├── root.go         # [MODIFY] Register push command (~10 lines)
│   └── push.go         # [CREATE] Push command implementation (~200 lines)
├── cmd/push_test.go    # [CREATE] Command tests (~100 lines)
└── docs/
    └── push-help.txt   # [CREATE] Help text template (~40 lines)
```

## Implementation Steps

### Step 1: Create Push Command Structure
**File**: `cmd/push.go`
**Lines**: ~200

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
    Use:   "push IMAGE_NAME",
    Short: "Push an OCI image to the integrated Gitea registry",
    Long: `Push an OCI image to the integrated Gitea registry.

The push command uploads container images to the Gitea registry
at https://gitea.cnoe.localtest.me:8443/.

Examples:
  # Push an image with authentication
  idpbuilder push myapp:latest --username admin --password secret

  # Push with insecure TLS (self-signed certificates)
  idpbuilder push myapp:latest --username admin --password secret --insecure`,
    Args: cobra.ExactArgs(1),
    RunE: runPush,
}

// Command configuration
type pushConfig struct {
    imageName string
    // Placeholder for future flags (auth, TLS)
}

func init() {
    // Command will be registered in root.go
    // Future flags will be added in subsequent efforts
}

// runPush executes the push command
func runPush(cmd *cobra.Command, args []string) error {
    config := &pushConfig{
        imageName: args[0],
    }

    // Validate image name format
    if err := validateImageName(config.imageName); err != nil {
        return fmt.Errorf("invalid image name: %w", err)
    }

    // Log command execution (temporary until implementation)
    fmt.Printf("Pushing image: %s\n", config.imageName)
    fmt.Println("Note: Push functionality will be implemented in Phase 4")

    return nil
}

// validateImageName performs basic validation on the image name
func validateImageName(name string) error {
    if name == "" {
        return fmt.Errorf("image name cannot be empty")
    }

    // Basic validation - will be enhanced in Phase 3
    // Check for basic format: [registry/]namespace/name[:tag]
    // For now, just ensure non-empty

    return nil
}
```

### Step 2: Register Command in Root
**File**: `cmd/root.go`
**Lines**: ~10 (modification)

```go
// Add to init() function or command registration area:
func init() {
    // ... existing init code ...

    // Register push command
    rootCmd.AddCommand(pushCmd)
}
```

### Step 3: Create Command Tests
**File**: `cmd/push_test.go`
**Lines**: ~100

```go
package cmd

import (
    "bytes"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPushCommand(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        wantErr   bool
        errMsg    string
    }{
        {
            name:    "valid image name",
            args:    []string{"push", "myapp:latest"},
            wantErr: false,
        },
        {
            name:    "missing image name",
            args:    []string{"push"},
            wantErr: true,
            errMsg:  "requires exactly 1 arg(s)",
        },
        {
            name:    "too many arguments",
            args:    []string{"push", "image1", "image2"},
            wantErr: true,
            errMsg:  "requires exactly 1 arg(s)",
        },
        {
            name:    "empty image name",
            args:    []string{"push", ""},
            wantErr: true,
            errMsg:  "image name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := rootCmd
            buf := new(bytes.Buffer)
            cmd.SetOut(buf)
            cmd.SetErr(buf)
            cmd.SetArgs(tt.args)

            err := cmd.Execute()

            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestValidateImageName(t *testing.T) {
    tests := []struct {
        name      string
        imageName string
        wantErr   bool
    }{
        {
            name:      "valid simple name",
            imageName: "myapp",
            wantErr:   false,
        },
        {
            name:      "valid with tag",
            imageName: "myapp:latest",
            wantErr:   false,
        },
        {
            name:      "valid with namespace",
            imageName: "namespace/myapp:v1.0",
            wantErr:   false,
        },
        {
            name:      "empty name",
            imageName: "",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateImageName(tt.imageName)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Step 4: Create Help Documentation
**File**: `docs/push-help.txt`
**Lines**: ~40

```
IDPBuilder Push Command

Usage:
  idpbuilder push IMAGE_NAME [flags]

Description:
  Push an OCI image to the integrated Gitea registry at
  https://gitea.cnoe.localtest.me:8443/

Arguments:
  IMAGE_NAME    The name of the image to push
                Format: [namespace/]name[:tag]
                Examples: myapp:latest, team/myapp:v1.0

Flags:
  -h, --help    help for push

Global Flags:
  [inherited from root command]

Examples:
  # Basic push (will add auth flags in effort 1.1.2)
  idpbuilder push myapp:latest

  # Push with specific tag
  idpbuilder push myapp:v1.0.0

  # Push with namespace
  idpbuilder push myteam/myapp:latest

Exit Codes:
  0    Success
  1    General error
  2    Invalid arguments

Notes:
  - Authentication flags will be added in effort 1.1.2
  - TLS configuration will be added in effort 1.1.3
  - Full push functionality will be implemented in Phase 4
```

## Size Management
- **Estimated Lines**: ~350 lines total
  - cmd/push.go: ~200 lines
  - cmd/push_test.go: ~100 lines
  - docs/push-help.txt: ~40 lines
  - cmd/root.go modification: ~10 lines
- **Measurement Tool**: ${PROJECT_ROOT}/tools/line-counter.sh
- **Check Frequency**: After each file completion
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Test Requirements
- **Unit Tests**: 90% coverage target
  - Command registration test
  - Argument validation tests
  - Image name validation tests
  - Help text display test
- **Integration Tests**: Not required for skeleton
- **E2E Tests**: Not required for skeleton
- **Test Files**:
  - cmd/push_test.go (primary test file)

## Pattern Compliance
- **Cobra Patterns**: Follow existing IDPBuilder command structure
- **Error Handling**: Use fmt.Errorf with error wrapping
- **Logging**: Use existing logging framework when available
- **Naming**: Follow Go naming conventions

## Dependencies and Integration Points
- **Dependencies**: None (using existing cobra)
- **Integration with 1.1.2**: Auth flags will extend pushCmd
- **Integration with 1.1.3**: TLS flags will extend pushCmd
- **Integration with Phase 4**: runPush will call push orchestrator

## Success Criteria
- ✅ Push command registered and appears in help
- ✅ Command accepts exactly one argument (image name)
- ✅ Basic validation of image name
- ✅ All tests passing with >90% coverage
- ✅ Help text is clear and comprehensive
- ✅ Code follows project patterns

## Risk Mitigation
1. **Risk**: Command conflicts with existing commands
   - **Mitigation**: Review existing commands first
   - **Validation**: Test command registration

2. **Risk**: Image name validation too restrictive
   - **Mitigation**: Keep validation minimal in skeleton
   - **Enhancement**: Full validation in Phase 3

## Implementation Order
1. Create cmd/push.go with basic structure
2. Add command registration to cmd/root.go
3. Write tests in cmd/push_test.go
4. Create help documentation
5. Run tests and verify coverage
6. Measure with line-counter.sh

## Notes for SW Engineer
- Start with test-first approach: write tests before implementation
- Keep implementation simple - this is just the skeleton
- Don't implement actual push logic - that's Phase 4
- Ensure command follows existing IDPBuilder patterns
- Validate that command appears in `idpbuilder --help`