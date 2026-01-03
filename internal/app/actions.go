package app

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type KeyAction struct {
	ShouldQuit    bool
	SetScreen     int
	ScreenChanged bool
	ClearLogs     bool
	ShowConfirm   bool
	StartMail     bool
	StopMail      bool
	CancelConfirm bool
	FileSelected  int
	UpdateDelay   int
	BlurInput     bool
	FocusInput    bool
	ImportFile    string
	ToggleHelp    bool
	ScrollUp      bool
	ScrollDown    bool
}

func (a *App) HandleKeyPress(key string, currentScreen int, confirmStart bool, mailStarted bool, inputFocused bool, selectedFile int, files []string, delayValue string) KeyAction {
	action := KeyAction{SetScreen: -1, FileSelected: -1}

	switch key {
	case "q", "ctrl+c":
		action.ShouldQuit = true

	case "1", "l":
		if currentScreen != 2 || !inputFocused {
			action.SetScreen = 0
			action.ScreenChanged = true
			action.BlurInput = true
		}

	case "2", "s":
		if currentScreen != 2 || !inputFocused {
			action.SetScreen = 1
			action.ScreenChanged = true
			action.BlurInput = true
		}

	case "3", "c":
		if currentScreen != 2 || !inputFocused {
			action.SetScreen = 2
			action.ScreenChanged = true
		}

	case "4", "i":
		action.SetScreen = 3
		action.ScreenChanged = true
		action.BlurInput = true

	case "5", "p":
		action.SetScreen = 4
		action.ScreenChanged = true
		action.BlurInput = true

	case "r", "R":
		if currentScreen == 0 {
			action.ClearLogs = true
		}

	case "b", "B":
		if currentScreen != 2 || !inputFocused {
			if !mailStarted && !confirmStart {
				action.ShowConfirm = true
			}
		}

	case "a":
		if currentScreen != 2 || !inputFocused {
			if mailStarted {
				action.StopMail = true
			}
		}

	case "y":
		if confirmStart {
			action.StartMail = true
		}

	case "n":
		if confirmStart {
			action.CancelConfirm = true
		}

	case "up":
		if currentScreen == 3 && len(files) > 0 && selectedFile > 0 {
			action.FileSelected = selectedFile - 1
		} else if currentScreen == 0 {
			action.ScrollUp = true
		}

	case "down":
		if currentScreen == 3 && len(files) > 0 && selectedFile < len(files)-1 {
			action.FileSelected = selectedFile + 1
		} else if currentScreen == 0 {
			action.ScrollDown = true
		}

	case "enter":
		if currentScreen == 2 {
			if inputFocused {
				if val, err := strconv.Atoi(delayValue); err == nil && val > 0 {
					action.UpdateDelay = val
				}
				action.BlurInput = true
			} else {
				action.FocusInput = true
			}
		} else if currentScreen == 3 && len(files) > 0 && selectedFile >= 0 && selectedFile < len(files) {
			action.ImportFile = files[selectedFile]
		}

	case "esc":
		action.BlurInput = true

	case "h":
		action.ToggleHelp = true
	}

	return action
}

func (a *App) ImportEmailsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(string(data), -1)
	a.addLog(fmt.Sprintf("Found %d emails in %s", len(emails), filename))

	currentData, _ := os.ReadFile(a.cfg.Database.Path)
	currentContent := string(currentData)

	f, err := os.OpenFile(a.cfg.Database.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	added := 0
	for _, email := range emails {
		if !strings.Contains(currentContent, email+" PENDING") &&
			!strings.Contains(currentContent, email+" DONE") &&
			!strings.Contains(currentContent, email+" FAILED") {
			f.WriteString(email + " PENDING\n")
			added++
		}
	}

	a.addLog(fmt.Sprintf("Imported %d emails from %s", added, filename))
	a.updateStats()
	a.addLog(fmt.Sprintf("New pending count: %d", a.stats.Pending))

	return nil
}

func (a *App) ClearLogs() {
	a.mu.Lock()
	a.logs = []string{"Logs cleared"}
	a.mu.Unlock()
}
