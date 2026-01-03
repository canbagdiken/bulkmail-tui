package app

import (
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/fsnotify/fsnotify"
)

// Email status constants
const (
	StatusPending      = "PENDING"
	StatusSending      = "SENDING"
	StatusDone         = "DONE"
	StatusFailed       = "FAILED"
	StatusUnsubscribed = "UNSUBSCRIBED"
)

// Screen constants
const (
	ScreenLogs = iota
	ScreenStats
	ScreenPreferences
	ScreenImport
	ScreenPending
)

type tickMsg time.Time

type Config struct {
	SMTP struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		FromEmail string `yaml:"from_email"`
		FromName  string `yaml:"from_name"`
	} `yaml:"smtp"`

	Mail struct {
		DelaySeconds int    `yaml:"delay_seconds"`
		Subject      string `yaml:"subject"`
		Template     string `yaml:"template"`
	} `yaml:"mail"`

	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
}

type Recipient struct {
	Email  string
	Status string
	Error  string
}

type Stats struct {
	Total        int
	Sending      int
	Pending      int
	Sent         int
	Failed       int
	Unsubscribed int
}

type ViewData struct {
	StatusText     string
	IsRunning      bool
	PendingCount   int
	PendingEmails  []PendingEmail
	CurrentScreen  int
	Logs           []string
	Stats          Stats
	DelaySeconds   int
	ImportFiles    []string
	SelectedFile   int
	TabNames       []string
	LogsContent    string
	StatsContent   string
	PendingContent string
	ImportContent  string
}

type PendingEmail struct {
	Email     string
	IsSending bool
}

type Job struct {
	Email string
}

// App represents the application state
type App struct {
	cfg            *Config
	htmlBody       []byte
	mu             sync.Mutex
	viewDataMu     sync.Mutex
	stopCh         chan bool
	delaySeconds   int
	booted         bool
	stats          Stats
	logs           []string
	Watcher        *fsnotify.Watcher
	viewData       ViewData
	noPendingCount int
}

type keyMap struct {
	Quit        key.Binding
	Logs        key.Binding
	Stats       key.Binding
	Preferences key.Binding
	Import      key.Binding
	Pending     key.Binding
	Boot        key.Binding
	Stop        key.Binding
	Up          key.Binding
	Down        key.Binding
	Clear       key.Binding
	Help        key.Binding
}

const tabLineText = "Logs | Stats | Preferences | Import | Pending"

const (
	colorTabLine       = "5"
	colorBackground    = "0"
	viewportHeightPad  = 4
	defaultWidth       = 80
	defaultHeight      = 20
	textInputCharLimit = 10
	textInputWidth     = 20
)

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Logs, k.Stats, k.Preferences},
		{k.Import, k.Pending, k.Boot, k.Stop},
		{k.Up, k.Down, k.Clear, k.Help},
	}
}
