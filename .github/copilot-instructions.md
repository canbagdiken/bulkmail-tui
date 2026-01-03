# GitHub Copilot Instructions for BulkMail TUI

## Project Overview

BulkMail TUI is a Terminal User Interface application for sending bulk emails via SMTP. Built with Go 1.20+ and Bubble Tea framework.

## Architecture

### Core Components

1. **app.go** - Application state management, dispatcher, stats tracking
2. **tui.go** - Terminal UI using Bubble Tea framework
3. **mail.go** - SMTP email sending with TLS support
4. **database.go** - Semicolon-delimited text file database operations
5. **types.go** - Type definitions and structs
6. **config.go** - YAML configuration management
7. **actions.go** - Import and utility functions

### Key Patterns

- **ViewData Caching**: All UI display data is pre-computed in `updateViewData()` to avoid runtime calculations
- **Mutex Locking**: Separate locks for logs and viewData to prevent deadlocks
- **Smart Delay**: Calculates time since last email, only waits for remaining duration
- **Status State Machine**: PENDING → SENDING → DONE/FAILED
- **Auto-recovery**: Resets stuck SENDING emails (>5min) to PENDING on startup

## Database Format

```
{ISO8601_DATE} ; {STATUS} ; {EMAIL} ; {ERROR_MESSAGE}
```

**Status Constants:**
- `StatusPending = "PENDING"`
- `StatusSending = "SENDING"`
- `StatusDone = "DONE"`
- `StatusFailed = "FAILED"`
- `StatusUnsubscribed = "UNSUBSCRIBED"`

## Coding Style

### Go Conventions

- Use standard Go formatting (`gofmt`)
- Error handling: always check and propagate errors
- Context: use `context.Context` for timeouts and cancellations
- Mutexes: acquire locks as late as possible, release as early as possible

### Comments

- Write comments in **English only**
- Document all exported functions and types
- Explain complex logic with inline comments
- Use format: `// FunctionName: Brief description of what it does`

### Naming

- Use camelCase for variables and functions
- Use PascalCase for exported types and functions
- Use ALL_CAPS for constants
- Prefix unexported functions with lowercase

### Error Messages

- Start with lowercase (except proper nouns)
- Be descriptive but concise
- Use `fmt.Errorf()` with `%w` for error wrapping
- Example: `fmt.Errorf("failed to load config: %w", err)`

## Common Tasks

### Adding a New Email Status

1. Add constant to `database.go`: `const StatusNewStatus = "NEWSTATUS"`
2. Update `GetStats()` switch statement
3. Update `GetPendingEmails()` if needed
4. Add color code in `app.go` `updateViewData()` if UI display needed

### Adding a New UI Tab

1. Add tab name to `viewData.TabNames` in `app.go`
2. Update keyboard shortcuts in `tui.go` `Update()` method
3. Add case in `tui.go` `getContent()` method
4. Update help text in `View()` method

### Modifying Email Template

- Use `{{email}}` placeholder for recipient email
- HTML is automatically converted to plain text
- Headers are automatically added (Message-ID, Date, Reply-To, etc.)

### Database Operations

- Always use provided functions in `database.go`
- Don't access file directly
- Use `UpdateStatus()` for atomic status changes
- Use `GetNextPending()` to get and mark emails as SENDING atomically

## Dependencies

```go
github.com/charmbracelet/bubbles // UI components
github.com/charmbracelet/bubbletea // TUI framework
github.com/charmbracelet/lipgloss // Styling
gopkg.in/gomail.v2 // SMTP client
gopkg.in/yaml.v3 // YAML parsing
github.com/fsnotify/fsnotify // File watching
```

## Testing Considerations

- Test with small delay values (e.g., 5 seconds) during development
- Use example emails: `test@example.com`, `user@example.org`
- Never commit real SMTP credentials
- Always test on your own email before bulk sending

## Security Guidelines

- **Never hardcode credentials** - use config.yaml (gitignored)
- **Never log passwords** - even in debug mode
- **Sanitize email addresses** - validate format before sending
- **Use TLS** - always enable for port 587, SSL for port 465
- **Rate limiting** - respect SMTP provider limits with appropriate delays

## Performance Tips

- **ViewData Pattern**: Pre-compute all UI data, don't calculate in render loop
- **Mutex Efficiency**: Keep critical sections small
- **Smart Delay**: Calculate elapsed time since last send
- **Goroutines**: Use channels for communication, avoid shared memory when possible
- **File I/O**: Batch operations when possible

## Common Pitfalls to Avoid

1. **Mutex Deadlock**: Don't call a function that locks mutex from inside another locked section
2. **Goroutine Leaks**: Always ensure goroutines can exit (use context cancellation)
3. **File Handle Leaks**: Always `defer file.Close()` after opening
4. **Race Conditions**: Use mutexes for all shared state access
5. **Timezone Issues**: Always use UTC for database timestamps

## ANSI Color Codes

Used in logs for visual feedback:

```go
"\033[32m" // Green - Success
"\033[31m" // Red - Error
"\033[33m" // Yellow - Warning
"\033[36m" // Cyan - Info
"\033[34m" // Blue - Action
"\033[35m" // Magenta - Status change
"\033[90m" // Gray - Debug
"\033[0m"  // Reset
```

## File Structure Conventions

```
bulkmail-tui/
├── .github/
│   └── workflows/
│       └── release.yml      # Automated releases
├── *.go                      # Go source files
├── config.yaml              # User config (gitignored)
├── config.yaml.example      # Example config
├── data.txt                 # Email database (gitignored)
├── mail.html                # Email template
├── go.mod                   # Go dependencies
├── README.md                # User documentation
├── LICENSE                  # MIT License
└── .gitignore              # Git ignore rules
```

## When Making Changes

1. **Preserve existing patterns** - don't introduce new paradigms without discussion
2. **Maintain backwards compatibility** - config and database format should not break
3. **Update README** - if adding user-facing features
4. **Test thoroughly** - especially SMTP and database operations
5. **Check for race conditions** - run with `-race` flag during development

## AI Assistant Specific Notes

- Prefer reading entire functions rather than snippets for context
- When modifying mutex-protected code, review all lock/unlock pairs
- When adding new goroutines, ensure proper cleanup mechanism
- When dealing with email operations, consider error cases (network timeout, SMTP rejection)
- Always check if a similar function exists before creating a new one
- Respect the ViewData caching pattern - UI should read from cache, not compute
