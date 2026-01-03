package app

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	watcherSetupDelay  = 100 * time.Millisecond
	dispatcherInterval = 1 * time.Second
)

// Init: Initializes the application, loads config, sets up watcher
func (a *App) Init() error {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	a.cfg = cfg

	htmlBody, err := os.ReadFile(cfg.Mail.Template)
	if err != nil {
		return fmt.Errorf("failed to load template: %v", err)
	}
	a.htmlBody = htmlBody

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	a.Watcher = watcher

	// Ensure data file exists
	if err := InitDB(cfg.Database.Path); err != nil {
		return fmt.Errorf("failed to init db: %v", err)
	}

	err = watcher.Add(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to watch file: %v", err)
	}
	time.Sleep(watcherSetupDelay)

	a.stopCh = make(chan bool, 1)
	a.delaySeconds = cfg.Mail.DelaySeconds
	a.booted = false
	a.logs = []string{"BulkMail TUI started...", "Initializing database...", "Setting up watcher...", "Loading configuration..."}

	// Initialize viewData
	a.viewData.TabNames = []string{"Logs", "Stats", "Preferences", "Import", "Pending"}
	a.viewData.DelaySeconds = a.delaySeconds
	a.viewData.IsRunning = false
	a.viewData.StatusText = "STOPPED"

	a.addLog("Updating initial stats...")
	a.updateStats()

	a.addLog("Starting dispatcher...")
	a.startDispatcher()

	a.addLog("Initialization complete!")

	return nil
}

func (a *App) updateStats() {
	stats, err := GetStats(a.cfg.Database.Path)
	if err == nil {
		a.mu.Lock()
		a.stats = *stats
		a.mu.Unlock()
		a.updateViewData()
	}
}

func (a *App) updateViewData() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Calculate pending count
	a.viewData.PendingCount = a.stats.Pending

	// Set status
	a.viewData.IsRunning = a.booted
	if a.booted {
		a.viewData.StatusText = "RUNNING"
	} else {
		a.viewData.StatusText = "STOPPED"
	}

	// Copy stats
	a.viewData.Stats = a.stats

	// Copy delay
	a.viewData.DelaySeconds = a.delaySeconds

	// Copy logs
	a.viewData.Logs = make([]string, len(a.logs))
	copy(a.viewData.Logs, a.logs)

	// Load pending emails
	pendingEmails, err := GetPendingEmails(a.cfg.Database.Path)
	if err != nil {
		a.viewData.PendingEmails = []PendingEmail{}
	} else {
		a.viewData.PendingEmails = pendingEmails
	}

	// Prepare tab names
	a.viewData.TabNames = []string{"Logs", "Stats", "Preferences", "Import", "Pending"}

	// Prepare logs content with colors
	logs := a.viewData.Logs
	if len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	}
	a.viewData.LogsContent = ""
	for _, log := range logs {
		originalLog := log
		if len(log) > 80 {
			log = log[:80] + "..."
		}

		// Apply colors based on log content
		coloredLog := log
		if strings.HasPrefix(originalLog, "✓") {
			// Success - Green
			coloredLog = "\033[32m" + log + "\033[0m"
		} else if strings.Contains(strings.ToLower(originalLog), "error") || strings.HasPrefix(originalLog, "Error") {
			// Error - Red
			coloredLog = "\033[31m" + log + "\033[0m"
		} else if strings.Contains(originalLog, "waiting") || strings.Contains(originalLog, "Wait") {
			// Waiting - Yellow
			coloredLog = "\033[33m" + log + "\033[0m"
		} else if strings.HasPrefix(originalLog, "DEBUG:") {
			// Debug - Gray
			coloredLog = "\033[90m" + log + "\033[0m"
		} else if strings.Contains(originalLog, "Checking") || strings.Contains(originalLog, "Found") {
			// Info - Cyan
			coloredLog = "\033[36m" + log + "\033[0m"
		} else if strings.Contains(originalLog, "Sending email") {
			// Sending - Blue
			coloredLog = "\033[34m" + log + "\033[0m"
		} else if strings.Contains(originalLog, "STOPPED") || strings.Contains(originalLog, "stopped") {
			// Stopped - Magenta
			coloredLog = "\033[35m" + log + "\033[0m"
		}

		a.viewData.LogsContent += coloredLog + "\n"
	}

	// Prepare stats content
	a.viewData.StatsContent = fmt.Sprintf("Statistics:\nTotal: %d\nPending: %d\nSending: %d\nSent: %d\nFailed: %d\n",
		a.viewData.Stats.Total,
		a.viewData.Stats.Pending,
		a.viewData.Stats.Sending,
		a.viewData.Stats.Sent,
		a.viewData.Stats.Failed)

	// Prepare pending content
	a.viewData.PendingContent = "Pending Emails:\n\n"
	for _, pendingEmail := range a.viewData.PendingEmails {
		if pendingEmail.IsSending {
			a.viewData.PendingContent += "⏳ " + pendingEmail.Email + "\n"
		} else {
			a.viewData.PendingContent += "📧 " + pendingEmail.Email + "\n"
		}
	}
}

func (a *App) addLog(log string) {
	a.mu.Lock()
	a.logs = append(a.logs, log)
	a.mu.Unlock()

	// Update viewData after adding log
	a.updateViewData()
}

func (a *App) updateLastLog(log string) {
	a.mu.Lock()
	if len(a.logs) > 0 {
		a.logs[len(a.logs)-1] = log
	} else {
		a.logs = append(a.logs, log)
	}
	a.mu.Unlock()

	// Update viewData after updating log
	a.updateViewData()
}

func (a *App) startDispatcher() {

	go func() {
		a.addLog("Dispatcher started")

		// Reset stuck SENDING records on startup
		count, err := ResetStuckSending(a.cfg.Database.Path, 5*time.Minute)
		if err != nil {
			a.addLog(fmt.Sprintf("ResetStuckSending error: %v", err))
		} else if count > 0 {
			a.addLog(fmt.Sprintf("Reset %d stuck SENDING records to PENDING", count))
			a.updateStats()
		}

		ticker := time.NewTicker(dispatcherInterval)
		defer ticker.Stop()

		for {
			select {
			case <-a.stopCh:
				a.addLog("Dispatcher stopped")
				return
			case <-ticker.C:
				a.mu.Lock()
				b := a.booted
				a.mu.Unlock()

				if !b {
					continue
				}

				a.addLog("Checking for pending emails...")
				recipient, err := GetNextPending(a.cfg.Database.Path)
				if err != nil {
					if !errors.Is(err, ErrNoPendingRecipients) {
						a.updateLastLog(fmt.Sprintf("GetNextPending error: %v", err))
						a.noPendingCount = 0
					} else {
						a.noPendingCount++
						a.updateLastLog(fmt.Sprintf("No pending emails found (%d/3)", a.noPendingCount))
						if a.noPendingCount >= 3 {
							a.mu.Lock()
							a.booted = false
							a.mu.Unlock()
							a.noPendingCount = 0
							a.addLog("No pending emails after 3 checks, switching to STOPPED")
							a.updateStats()
						}
					}
					continue
				}
				if recipient == nil {
					a.noPendingCount++
					a.updateLastLog(fmt.Sprintf("No pending emails found (%d/3)", a.noPendingCount))
					if a.noPendingCount >= 3 {
						a.mu.Lock()
						a.booted = false
						a.mu.Unlock()
						a.noPendingCount = 0
						a.addLog("No pending emails after 3 checks, switching to STOPPED")
						a.updateStats()
					}
					continue
				}

				a.noPendingCount = 0
				a.updateLastLog(fmt.Sprintf("Found pending email: %s", recipient.Email))

				// Check last sent time and calculate delay
				lastSentTime, err := GetLastSentTime(a.cfg.Database.Path)
				if err == nil && !lastSentTime.IsZero() {
					elapsed := time.Since(lastSentTime)
					delay := time.Duration(a.delaySeconds) * time.Second

					if elapsed < delay {
						waitTime := delay - elapsed
						a.addLog(fmt.Sprintf("Last email sent %.0f seconds ago, waiting %.0f more seconds...", elapsed.Seconds(), waitTime.Seconds()))
						time.Sleep(waitTime)
					} else {
						a.addLog(fmt.Sprintf("Last email sent %.0f seconds ago, proceeding immediately", elapsed.Seconds()))
					}
				} else {
					a.addLog("No previous emails sent, proceeding immediately")
				}

				a.addLog(fmt.Sprintf("Sending email to %s...", recipient.Email))
				err = SendMail(a.cfg, recipient.Email, a.cfg.Mail.Subject, string(a.htmlBody))

				if err != nil {
					a.addLog(fmt.Sprintf("Error sending to %s: %v", recipient.Email, err))
					if updateErr := UpdateStatus(a.cfg.Database.Path, recipient.Email, StatusFailed, err.Error()); updateErr != nil {
						a.addLog(fmt.Sprintf("UpdateStatus error: %v", updateErr))
					}
				} else {
					a.addLog(fmt.Sprintf("✓ Sent to %s", recipient.Email))
					if updateErr := UpdateStatus(a.cfg.Database.Path, recipient.Email, StatusDone, ""); updateErr != nil {
						a.addLog(fmt.Sprintf("UpdateStatus error: %v", updateErr))
					}
				}
				a.updateStats()
			case event := <-a.Watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					a.addLog("Database file changed, updating...")
					if err := a.UpdateDataFile(a.cfg.Database.Path); err != nil {
						a.addLog(fmt.Sprintf("UpdateDataFile error: %v", err))
					}
					a.updateStats()
				}
			case err := <-a.Watcher.Errors:
				log := fmt.Sprintf("Watcher error: %v", err)
				a.addLog(log)
			}
		}
	}()
}

func (a *App) UpdateDataFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	changed := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, ";") && strings.Contains(line, "@") {
			lines[i] = "0000-00-00T00:00:00Z ; " + StatusPending + " ; " + line
			changed = true
			a.addLog(fmt.Sprintf("Converted: %s", line))
		}
	}
	if changed {
		a.addLog("Data file updated with new pending entries.")
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	}
	return nil
}

