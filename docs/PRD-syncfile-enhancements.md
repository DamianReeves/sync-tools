# Product Requirements Document: SyncFile Post-Sync Actions

## Executive Summary

This PRD defines post-sync actions for SyncFile format, enabling automated patch application, custom scripts, and file transformations after synchronization operations. This extends sync-tools from simple file synchronization to a comprehensive deployment and workflow automation platform.

## Problem Statement

Current SyncFile capabilities are limited to synchronization operations:
- No way to apply patches or configuration changes after sync completion
- Cannot execute custom scripts or transformations as part of sync workflow
- No support for conditional actions based on sync results
- Limited ability to create complex deployment workflows that require post-sync processing
- No mechanism for applying inline patches or external patch files

## Target Personas

1. **DevOps Engineers**: Deploy applications with configuration patches and post-deployment scripts
2. **System Administrators**: Sync files and apply environment-specific customizations
3. **Developers**: Automate development environment setup with post-sync configurations
4. **Release Engineers**: Complex deployment workflows with validation and rollback capabilities
5. **Data Engineers**: Sync datasets and apply transformations or schema updates

## Solution Overview

Extend SyncFile with post-sync action instructions that execute after successful synchronization:

- **PATCH**: Apply git-style patches (inline or from files)
- **SCRIPT**: Execute custom scripts with sync context
- **TRANSFORM**: Apply file transformations using templates or rules  
- **VALIDATE**: Run validation checks with rollback capabilities
- **NOTIFY**: Send notifications about sync completion

## Detailed Requirements

### Core Post-Sync Action Instructions

#### 1. PATCH Instruction

Apply git-style patches after synchronization completes.

**Syntax:**
```dockerfile
PATCH [OPTIONS] <patch-source>
PATCH --file changes.patch

# Full git patch format (for complex multi-file patches)
PATCH:
  diff --git a/config/app.yml b/config/app.yml
  index 1234567..abcdefg 100644
  --- a/config/app.yml
  +++ b/config/app.yml
  @@ -10,7 +10,7 @@ database:
     host: localhost
  -  port: 5432
  +  port: 5433
     name: myapp
END PATCH

# Concise formats (recommended for simple changes)

# Minimal unified diff (no git metadata)
PATCH config/app.yml:
  @@ -10,3 +10,3 @@
     host: localhost
  -  port: 5432
  +  port: 5433
     name: myapp
END PATCH

```

**Options:**
- `--file <path>`: Apply patch from external file
- `:` (colon): Start inline patch definition (ends with `END PATCH`)
- `--target <directory>`: Apply patch to specific directory (default: sync destination)
- `--dry-run`: Show what patch would do without applying
- `--backup`: Create backup before applying patch
- `--force`: Apply patch even if there are conflicts (with .orig files)

**Patch Format Support:**
- **Full Git Patch**: Complete git diff format with all metadata (for complex multi-file patches)
- **Minimal Diff**: Unified diff without git metadata (`@@` context lines specify location)

#### 2. SCRIPT Instruction

Execute custom scripts with access to sync context and results.

**Syntax:**
```dockerfile
SCRIPT [OPTIONS] <script-command>
SCRIPT --file deploy.sh
SCRIPT:
  #!/bin/bash
  echo "Sync completed: $SYNC_FILES_CHANGED files changed"
  systemctl reload nginx
END SCRIPT
```

**Options:**
- `--file <path>`: Execute script from file
- `:` (colon): Start inline script definition (ends with `END SCRIPT`)
- `--working-dir <path>`: Set working directory for script execution
- `--timeout <seconds>`: Script execution timeout (default: 300)
- `--on-failure <action>`: Action on script failure (abort, continue, rollback)
- `--env <key=value>`: Additional environment variables

**Environment Variables Available to Scripts:**
- `SYNC_SOURCE`: Source directory path
- `SYNC_DEST`: Destination directory path  
- `SYNC_MODE`: Sync mode (one-way, two-way)
- `SYNC_FILES_CHANGED`: Number of files modified
- `SYNC_FILES_CREATED`: Number of files created
- `SYNC_FILES_DELETED`: Number of files deleted
- `SYNC_TOTAL_SIZE`: Total bytes transferred
- `SYNC_DURATION`: Sync operation duration in seconds
- `SYNC_SUCCESS`: "true" if sync completed without errors

#### 3. REPLACE Instruction

Perform text replacements and content substitutions in files.

**Syntax:**
```dockerfile
REPLACE [OPTIONS] <target-file> <replacement-spec>
REPLACE --file config.yml --find "old-value" --replace "new-value"

# sed-style replacement
REPLACE config/app.yml:
  s/port: 5432/port: 5433/g
  s/localhost/${DB_HOST}/g
END REPLACE

# Block replacement (for multi-line changes)
REPLACE config/app.yml:
  FIND:
    database:
      host: localhost
      port: 5432
  REPLACE:
    database:
      host: ${DB_HOST}
      port: ${DB_PORT}
END REPLACE

# Multiple replacements in one file
REPLACE config/app.yml:
  s/environment: development/environment: ${DEPLOY_ENV}/
  s/debug: true/debug: false/
  FIND:
    old_section:
      - item1
      - item2
  REPLACE:
    new_section:
      - ${ITEM1}
      - ${ITEM2}
END REPLACE
```

**Options:**
- `--file <path>`: Apply replacements to external file
- `:` (colon): Start inline replacement definition (ends with `END REPLACE`)
- `--backup`: Create backup before applying replacements
- `--target <directory>`: Apply replacements in specific directory (default: sync destination)

**Replacement Format Support:**
- **sed-style**: Regular expression replacements (`s/pattern/replacement/flags`)
- **Block Find/Replace**: `FIND:` ... `REPLACE:` blocks (best for multi-line content)
- **Mixed**: Combine sed-style and block replacements in single instruction

#### 4. TRANSFORM Instruction

Apply file transformations using templates or custom processors.

**Syntax:**
```dockerfile
TRANSFORM [OPTIONS] <target-pattern> <transformation>
TRANSFORM --template config/*.yml.tmpl
TRANSFORM --processor custom-config-processor config/
```

**Options:**
- `--template <pattern>`: Process template files (removes .tmpl extension)
- `--processor <command>`: Run custom transformation processor
- `--backup`: Create backup before transformation
- `--target <directory>`: Directory to operate on (default: sync destination)

#### 5. VALIDATE Instruction

Run validation checks with rollback capabilities.

**Syntax:**
```dockerfile
VALIDATE [OPTIONS] <validation-command>
VALIDATE --command "nginx -t"
VALIDATE --script validate-config.sh
VALIDATE --rollback-on-failure
```

**Options:**
- `--command <cmd>`: Run validation command
- `--script <path>`: Run validation script
- `--rollback-on-failure`: Restore from backup if validation fails
- `--timeout <seconds>`: Validation timeout (default: 60)

#### 6. NOTIFY Instruction

Send notifications about sync completion and results.

**Syntax:**
```dockerfile
NOTIFY [OPTIONS] <message>
NOTIFY --slack "#deployments" "Production sync completed successfully"
NOTIFY --email "ops@company.com" --subject "Sync Report"
NOTIFY --webhook "https://api.monitoring.com/events"
```

### Advanced Post-Sync Features

#### Conditional Actions

Execute actions based on sync results or conditions:

```dockerfile
# Only apply patch if files were actually changed
IF SYNC_FILES_CHANGED > 0
  PATCH --file production.patch
  SCRIPT systemctl reload nginx
END IF

# Different actions based on sync mode
IF SYNC_MODE == "two-way"
  VALIDATE --command "check-conflicts.sh"
  NOTIFY --slack "Two-way sync completed with conflict resolution"
END IF
```

#### Action Groups and Dependencies

Group related actions and define dependencies:

```dockerfile
GROUP database-update
  PATCH --file schema-update.patch
  SCRIPT run-migrations.sh
  VALIDATE --command "test-database-connection.sh"
END GROUP

GROUP web-server  
  DEPENDS-ON database-update
  TRANSFORM --sed 's/old-version/new-version/g' version.txt
  SCRIPT restart-web-server.sh
  VALIDATE --command "curl -f http://localhost/health"
END GROUP
```

#### Rollback Capabilities

Automatic rollback on failure:

```dockerfile
# Enable automatic rollback for entire SyncFile
ROLLBACK-ON-FAILURE true

SYNC src/ dest/
PATCH --file config.patch --backup
SCRIPT --file deploy.sh --on-failure rollback
VALIDATE --command "test-deployment.sh" --rollback-on-failure
```

### Error Handling and Recovery

#### Failure Modes

- **abort**: Stop execution immediately (default)
- **continue**: Log error and continue with next action
- **rollback**: Restore from backups and undo changes
- **retry**: Retry action with exponential backoff

#### Backup Strategy

Automatic backup creation before destructive operations:

```dockerfile
# Configure backup strategy
BACKUP-STRATEGY incremental  # full, incremental, none
BACKUP-RETENTION 7d          # Keep backups for 7 days

SYNC src/ dest/
PATCH --file changes.patch   # Automatic backup created
TRANSFORM --sed 's/a/b/g' config.yml  # Automatic backup created
```

## Example SyncFiles with Post-Sync Actions

### 1. Application Deployment

```dockerfile
# Production application deployment
VAR APP_NAME=myapp
VAR DEPLOY_ENV=production

# Sync application files
SYNC /build/${APP_NAME} /var/www/${APP_NAME}
MODE one-way
EXCLUDE node_modules/
EXCLUDE *.log

# Apply environment-specific configuration
PATCH:
  diff --git a/config/app.yml b/config/app.yml
  --- a/config/app.yml
  +++ b/config/app.yml
  @@ -1,3 +1,3 @@
   environment: development
  -database_url: localhost:5432
  +database_url: ${DB_HOST}:${DB_PORT}
END PATCH

# Run database migrations
SCRIPT:
  #!/bin/bash
  cd /var/www/${APP_NAME}
  ./bin/migrate up
END SCRIPT

# Restart services
SCRIPT systemctl restart ${APP_NAME}
SCRIPT systemctl reload nginx

# Validate deployment
VALIDATE --command "curl -f http://localhost/${APP_NAME}/health"
VALIDATE --rollback-on-failure

# Notify team
NOTIFY --slack "#deployments" "✅ ${APP_NAME} deployed to ${DEPLOY_ENV}"
```

### 2. Development Environment Setup

```dockerfile
# Setup development environment
VAR PROJECT_NAME=awesome-project
VAR DEV_USER=${USER}

# Sync project template
SYNC /templates/project-template/ ./${PROJECT_NAME}/
MODE one-way

# Customize for user
TRANSFORM --sed "s/PROJECT_NAME/${PROJECT_NAME}/g" ./${PROJECT_NAME}/README.md
TRANSFORM --sed "s/DEV_USER/${DEV_USER}/g" ./${PROJECT_NAME}/.env.example

# Copy environment config
SCRIPT cp ./${PROJECT_NAME}/.env.example ./${PROJECT_NAME}/.env

# Install dependencies
SCRIPT --working-dir ./${PROJECT_NAME} npm install

# Initialize git repository
SCRIPT --working-dir ./${PROJECT_NAME} git init
SCRIPT --working-dir ./${PROJECT_NAME} git add .
SCRIPT --working-dir ./${PROJECT_NAME} git commit -m "Initial project setup"

# Validate setup
VALIDATE --command "cd ${PROJECT_NAME} && npm test"

NOTIFY --email "${DEV_USER}@company.com" "Development environment ready: ${PROJECT_NAME}"
```

### 3. Configuration Management

```dockerfile
# Multi-server configuration deployment
VAR CONFIG_VERSION=v2.1.0

# Sync base configuration
SYNC /config/base/ /etc/myapp/
MODE one-way
GITIGNORE true

# Apply server-specific patches
IF HOSTNAME == "web-01"
  PATCH --file patches/web-01.patch
ELIF HOSTNAME == "web-02"  
  PATCH --file patches/web-02.patch
END IF

# Transform templates with server-specific values
TRANSFORM --template /etc/myapp/*.yml.tmpl

# Validate configuration
VALIDATE --command "/usr/local/bin/validate-config /etc/myapp/"

# Reload services on successful validation
GROUP service-reload
  DEPENDS-ON validation
  SCRIPT systemctl reload myapp
  SCRIPT systemctl reload nginx
END GROUP

# Health check
VALIDATE --command "curl -f http://localhost:8080/health" --timeout 30

# Record deployment
SCRIPT echo "${CONFIG_VERSION} deployed at $(date)" >> /var/log/deployments.log
```

## Technical Implementation

### New Instruction Types

Add to `pkg/syncfile/syncfile.go`:

```go
const (
    // Post-sync action instructions
    InstPatch       InstructionType = "PATCH"       // PATCH --file changes.patch
    InstScript      InstructionType = "SCRIPT"      // SCRIPT deploy.sh
    InstReplace     InstructionType = "REPLACE"     // REPLACE config.yml s/old/new/g
    InstTransform   InstructionType = "TRANSFORM"   // TRANSFORM --template config/*.tmpl
    InstValidate    InstructionType = "VALIDATE"    // VALIDATE --command "test.sh"
    InstNotify      InstructionType = "NOTIFY"      // NOTIFY --slack "#ops" "Done"
    
    // Control flow instructions
    InstIf          InstructionType = "IF"          // IF condition
    InstElif        InstructionType = "ELIF"        // ELIF condition  
    InstElse        InstructionType = "ELSE"        // ELSE
    InstEndif       InstructionType = "ENDIF"       // ENDIF
    InstGroup       InstructionType = "GROUP"       // GROUP name
    InstDependsOn   InstructionType = "DEPENDS-ON"  // DEPENDS-ON group
    
    // Configuration instructions
    InstRollbackOnFailure InstructionType = "ROLLBACK-ON-FAILURE" // ROLLBACK-ON-FAILURE true
    InstBackupStrategy    InstructionType = "BACKUP-STRATEGY"     // BACKUP-STRATEGY incremental
    InstBackupRetention   InstructionType = "BACKUP-RETENTION"    // BACKUP-RETENTION 7d
)
```

### Action Execution Engine

New package `pkg/syncfile/actions` for post-sync action execution:

```go
type ActionExecutor interface {
    Execute(ctx context.Context, instruction Instruction, syncResult SyncResult) error
}

type PatchExecutor struct {
    workingDir string
    backupDir  string
}

type ScriptExecutor struct {
    workingDir string
    timeout    time.Duration
    env        map[string]string
}

type ReplaceExecutor struct {
    workingDir string
    backupDir  string
}
```

### Sync Result Context

Provide sync results to post-sync actions:

```go
type SyncResult struct {
    Source        string
    Destination   string
    Mode          string
    FilesChanged  int
    FilesCreated  int
    FilesDeleted  int
    TotalSize     int64
    Duration      time.Duration
    Success       bool
    Errors        []error
}
```

## Success Criteria

### Must-Have Features (MVP)
- [x] PATCH instruction with file and inline support
- [x] SCRIPT instruction with file and inline support  
- [x] Basic error handling (abort, continue modes)
- [x] Environment variable access to sync results
- [x] Backup creation for destructive operations

### Should-Have Features (V1.1)
- [ ] TRANSFORM instruction with template and sed support
- [ ] VALIDATE instruction with rollback capabilities
- [ ] Conditional execution (IF/ELIF/ELSE/ENDIF)
- [ ] Basic notification support (webhook, email)

### Could-Have Features (V2.0)
- [ ] Advanced control flow (GROUP, DEPENDS-ON)
- [ ] Full rollback capabilities with backup management
- [ ] Rich notification integrations (Slack, Teams, Discord)
- [ ] Custom processors and plugins
- [ ] Performance monitoring and metrics

## Risk Analysis

### Technical Risks
- **Complexity**: Post-sync actions significantly increase SyncFile complexity
  - *Mitigation*: Phased implementation starting with basic PATCH/SCRIPT
- **Security**: Executing arbitrary scripts poses security risks
  - *Mitigation*: Sandboxing, explicit enable flags, audit logging
- **Performance**: Large patches and scripts may slow deployment
  - *Mitigation*: Timeout controls, async execution options

### User Experience Risks  
- **Learning Curve**: New syntax may overwhelm existing users
  - *Mitigation*: Comprehensive documentation, examples, gradual rollout
- **Debugging**: Complex action chains difficult to troubleshoot
  - *Mitigation*: Detailed logging, dry-run modes, step-by-step execution

## Timeline

### Phase 1: Core Actions (4 weeks)
- Week 1-2: PATCH instruction implementation
- Week 3-4: SCRIPT instruction implementation  
- Testing and documentation

### Phase 2: Enhanced Features (3 weeks)
- Week 5-6: TRANSFORM and VALIDATE instructions
- Week 7: Error handling and backup strategies

### Phase 3: Advanced Features (3 weeks) 
- Week 8-9: Conditional execution and control flow
- Week 10: Notification system and polish

## Conclusion

Post-sync actions transform SyncFile from a simple synchronization format into a comprehensive deployment and automation platform. By starting with essential PATCH and SCRIPT instructions, we can deliver immediate value while building toward advanced workflow capabilities.

The phased approach allows for user feedback and iteration while managing complexity. The result will be a powerful yet approachable system for complex deployment and configuration management workflows.