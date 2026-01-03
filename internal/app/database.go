package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var ErrNoPendingRecipients = errors.New("no pending recipients")

// dbRecord represents a parsed database line
type dbRecord struct {
	Timestamp time.Time
	Status    string
	Email     string
	Error     string
}

// parseDBLine parses a database line into a dbRecord
func parseDBLine(line string) (*dbRecord, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, errors.New("empty line")
	}

	parts := strings.Split(line, ";")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid format: expected at least 3 parts, got %d", len(parts))
	}

	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
	if err != nil {
		// Fallback to zero time if parsing fails
		timestamp = time.Time{}
	}

	record := &dbRecord{
		Timestamp: timestamp,
		Status:    strings.TrimSpace(parts[1]),
		Email:     strings.TrimSpace(parts[2]),
	}

	if len(parts) >= 4 {
		record.Error = strings.TrimSpace(parts[3])
	}

	return record, nil
}

// formatDBLine formats a dbRecord into a database line
func (r *dbRecord) String() string {
	timestampStr := r.Timestamp.Format(time.RFC3339)
	if r.Timestamp.IsZero() {
		timestampStr = "0000-00-00T00:00:00Z"
	}

	line := timestampStr + " ; " + r.Status + " ; " + r.Email
	if r.Error != "" {
		line += " ; " + r.Error
	}
	return line
}

// Database wraps database file operations
type Database struct {
	path string
}

// NewDatabase creates a new Database instance
func NewDatabase(path string) *Database {
	return &Database{path: path}
}

// readLines reads all lines from database file
func (db *Database) readLines() ([]string, error) {
	file, err := os.Open(db.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes all lines to database file
func (db *Database) writeLines(lines []string) error {
	return os.WriteFile(db.path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// forEach iterates over all valid records
func (db *Database) forEach(fn func(record *dbRecord, lineIndex int) error) error {
	lines, err := db.readLines()
	if err != nil {
		return err
	}

	for i, line := range lines {
		record, err := parseDBLine(line)
		if err != nil {
			continue
		}
		if err := fn(record, i); err != nil {
			return err
		}
	}
	return nil
}

// updateRecord updates a specific record that matches the filter
func (db *Database) updateRecord(filter func(*dbRecord) bool, update func(*dbRecord)) error {
	lines, err := db.readLines()
	if err != nil {
		return err
	}

	for i, line := range lines {
		record, err := parseDBLine(line)
		if err != nil {
			continue
		}

		if filter(record) {
			update(record)
			lines[i] = record.String()
			break
		}
	}

	return db.writeLines(lines)
}

func GetNextPending(path string) (*Recipient, error) {
	db := NewDatabase(path)
	lines, err := db.readLines()
	if err != nil {
		return nil, err
	}

	for i, line := range lines {
		record, err := parseDBLine(line)
		if err != nil {
			continue
		}

		if record.Status == StatusPending {
			record.Timestamp = time.Now()
			record.Status = StatusSending
			lines[i] = record.String()

			if err := db.writeLines(lines); err != nil {
				return nil, err
			}

			return &Recipient{
				Email:  record.Email,
				Status: StatusPending,
			}, nil
		}
	}
	return nil, ErrNoPendingRecipients
}

func UpdateStatus(path, email, status, errorMsg string) error {
	db := NewDatabase(path)
	return db.updateRecord(
		func(r *dbRecord) bool { return r.Email == email },
		func(r *dbRecord) {
			r.Timestamp = time.Now()
			r.Status = status
			if errorMsg != "" {
				r.Error = strings.ReplaceAll(errorMsg, ";", ",")
			}
		},
	)
}

func GetStats(path string) (*Stats, error) {
	db := NewDatabase(path)
	stats := &Stats{}

	err := db.forEach(func(record *dbRecord, _ int) error {
		stats.Total++
		switch record.Status {
		case StatusPending:
			stats.Pending++
		case StatusSending:
			stats.Sending++
		case StatusDone:
			stats.Sent++
		case StatusFailed:
			stats.Failed++
		case StatusUnsubscribed:
			stats.Unsubscribed++
		}
		return nil
	})

	return stats, err
}

func GetLastSentTime(path string) (time.Time, error) {
	db := NewDatabase(path)
	var lastTime time.Time

	err := db.forEach(func(record *dbRecord, _ int) error {
		if (record.Status == StatusDone || record.Status == StatusFailed) && record.Timestamp.After(lastTime) {
			lastTime = record.Timestamp
		}
		return nil
	})

	return lastTime, err
}

func GetPendingEmails(path string) ([]PendingEmail, error) {
	db := NewDatabase(path)
	var pendingEmails []PendingEmail

	err := db.forEach(func(record *dbRecord, _ int) error {
		if record.Status == StatusPending || record.Status == StatusSending {
			pendingEmails = append(pendingEmails, PendingEmail{
				Email:     record.Email,
				IsSending: record.Status == StatusSending,
			})
		}
		return nil
	})

	return pendingEmails, err
}

// ResetStuckSending resets SENDING status to PENDING if older than timeout
func ResetStuckSending(path string, timeout time.Duration) (int, error) {
	db := NewDatabase(path)
	lines, err := db.readLines()
	if err != nil {
		return 0, err
	}

	count := 0
	now := time.Now()

	for i, line := range lines {
		record, err := parseDBLine(line)
		if err != nil {
			continue
		}

		if record.Status == StatusSending && !record.Timestamp.IsZero() {
			if now.Sub(record.Timestamp) > timeout {
				record.Status = StatusPending
				record.Error = "Reset from stuck SENDING state"
				lines[i] = record.String()
				count++
			}
		}
	}

	if count > 0 {
		err = db.writeLines(lines)
	}

	return count, err
}
