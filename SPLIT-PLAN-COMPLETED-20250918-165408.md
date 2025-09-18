# Split Plan for cli-commands Effort

## Current Situation
**Problem**: Entire codebase (10,147 lines) copied instead of focused CLI implementation
**Solution**: Complete reimplementation with proper scoping and splits

## Complete Split Inventory
**Total Expected Size**: ~1,500 lines (properly scoped CLI only)
**Splits Required**: 3
**Sole Planner**: Code Reviewer Agent

## Split Boundaries (NO OVERLAPS)

| Split | Description | Size | Files | Dependencies |
|-------|------------|------|-------|--------------|
| 001 | Core CLI Framework | 500 | root.go, helpers | None |
| 002 | Create/Delete Commands | 500 | create/, delete/ | Split 001 |
| 003 | Get/Version Commands | 500 | get/, version/ | Split 001 |

## Deduplication Matrix

| Component | Split 001 | Split 002 | Split 003 |
|-----------|-----------|-----------|-----------|
| Root command setup | ✅ | ❌ | ❌ |
| Command helpers | ✅ | ❌ | ❌ |
| Create command | ❌ | ✅ | ❌ |
| Delete command | ❌ | ✅ | ❌ |
| Get commands | ❌ | ❌ | ✅ |
| Version command | ❌ | ❌ | ✅ |

---

# SPLIT-PLAN-001.md
## Split 001 of 3: Core CLI Framework
**Planner**: Code Reviewer Agent
**Parent Effort**: cli-commands
**Branch**: phase2/wave2/cli-commands-split-001

### Boundaries
- **Previous Split**: None (first split)
- **This Split**: Split 001 of phase2/wave2/cli-commands
  - Path: efforts/phase2/wave2/cli-commands/split-001/
- **Next Split**: Split 002 of phase2/wave2/cli-commands
  - Path: efforts/phase2/wave2/cli-commands/split-002/

### Files in This Split
- pkg/cmd/root.go (50 lines) - Root command setup
- pkg/cmd/helpers/validation.go (100 lines) - Input validation
- pkg/cmd/helpers/output.go (100 lines) - Output formatting
- pkg/cmd/helpers/config.go (100 lines) - Configuration handling
- pkg/cmd/helpers/logger.go (100 lines) - Logging setup
- Tests: 50 lines

### Implementation Instructions
1. Create root command with cobra
2. Set up persistent flags (log-level, color output)
3. Implement validation helpers
4. Create output formatting utilities
5. Add configuration loading
6. Set up structured logging

### Size Target: 500 lines

---

# SPLIT-PLAN-002.md
## Split 002 of 3: Create/Delete Commands
**Planner**: Code Reviewer Agent
**Parent Effort**: cli-commands
**Branch**: phase2/wave2/cli-commands-split-002

### Boundaries
- **Previous Split**: Split 001 of phase2/wave2/cli-commands
  - Summary: Core CLI framework, helpers, validation
- **This Split**: Split 002 of phase2/wave2/cli-commands
  - Path: efforts/phase2/wave2/cli-commands/split-002/
- **Next Split**: Split 003 of phase2/wave2/cli-commands
  - Path: efforts/phase2/wave2/cli-commands/split-003/

### Files in This Split
- pkg/cmd/create/root.go (200 lines) - Create command implementation
- pkg/cmd/create/validate.go (50 lines) - Create validation
- pkg/cmd/delete/root.go (150 lines) - Delete command implementation
- pkg/cmd/delete/confirm.go (50 lines) - Deletion confirmation
- Tests: 50 lines

### Dependencies
- Requires Split 001 (imports helpers and root setup)

### Implementation Instructions
1. Import Split 001's helpers and root command
2. Implement create command with flags
3. Add validation for create inputs
4. Implement delete command with confirmation
5. Register commands with root
6. Add unit tests

### Size Target: 500 lines

---

# SPLIT-PLAN-003.md
## Split 003 of 3: Get/Version Commands
**Planner**: Code Reviewer Agent
**Parent Effort**: cli-commands
**Branch**: phase2/wave2/cli-commands-split-003

### Boundaries
- **Previous Split**: Split 002 of phase2/wave2/cli-commands
  - Summary: Create and Delete commands
- **This Split**: Split 003 of phase2/wave2/cli-commands
  - Path: efforts/phase2/wave2/cli-commands/split-003/
- **Next Split**: None (final split)

### Files in This Split
- pkg/cmd/get/root.go (50 lines) - Get subcommand root
- pkg/cmd/get/clusters.go (150 lines) - Get clusters command
- pkg/cmd/get/packages.go (100 lines) - Get packages command
- pkg/cmd/get/secrets.go (100 lines) - Get secrets command
- pkg/cmd/version/root.go (50 lines) - Version command
- Tests: 50 lines

### Dependencies
- Requires Split 001 (imports helpers and root setup)

### Implementation Instructions
1. Import Split 001's helpers and root command
2. Create get subcommand structure
3. Implement individual get commands
4. Add version command with build info
5. Register all commands with root
6. Add comprehensive tests

### Size Target: 500 lines

---

## Verification Checklist
- [x] No file appears in multiple splits
- [x] All CLI functionality covered
- [x] Each split compiles independently (with dependencies)
- [x] Dependencies properly ordered
- [x] Each split <800 lines (target ~500)
- [x] Clear boundaries between splits
- [x] No duplication of code