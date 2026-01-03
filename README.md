# BulkMail TUI

🚀 **Modern, lightweight bulk email sender with an interactive TUI**

A Terminal User Interface (TUI) application written in Go for sending bulk emails via SMTP. Track status, manage recipients, and monitor delivery in real-time.

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## ✨ Features

- 📧 **Bulk Email Sending** - Send emails to multiple recipients via SMTP
- 🎨 **Interactive TUI** - Beautiful terminal interface powered by Bubble Tea
- 📊 **Real-time Stats** - Track sent, failed, and pending emails
- 📝 **Status Tracking** - PENDING → SENDING → DONE/FAILED states
- 📂 **Smart Import** - Import emails from text files with regex extraction
- 🔄 **Auto-reload** - File watcher automatically detects changes
- ⚙️ **YAML Config** - Easy configuration management
- 🎯 **Template Support** - HTML email templates with placeholders
- 🚦 **Rate Limiting** - Configurable delay between sends
- 📦 **Single Binary** - No dependencies, just run

## 📸 Screenshots

```
┌─────────────────────────────────────────────────┐
│ Logs | Stats | Preferences | Import | Pending  │
├─────────────────────────────────────────────────┤
│ ✓ Sent to user1@example.com                    │
│ ✓ Sent to user2@example.com                    │
│ Database file changed, updating...             │
│ Imported 150 emails from contacts.txt          │
└─────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/canbagdiken/bulkmail-tui.git
cd bulkmail-tui

# Build
go build -o bulkmail .

# Run
./bulkmail
```

### Configuration

Create a `config.yaml` file in the project root:

```yaml
smtp:
  host: smtp.example.com
  port: 587
  username: your-email@example.com
  password: your-password-here
  from_email: your-email@example.com
  from_name: Your Name

mail:
  delay_seconds: 30
  subject: "Your Subject Here"
  template: mail.html
  num_workers: 1

database:
  path: data.txt
```

> **Note:** The application will create sample files if `config.yaml` or `data.txt` don't exist on first run.
  path: data.txt
```

### Usage

1. **Configure SMTP** - Edit `config.yaml` with your SMTP settings
2. **Add Recipients** - Add emails to `data.txt` (one per line) or import from files
3. **Run Application** - `./bulkmail`
4. **Boot System** - Press `B` to start sending
5. **Monitor** - Watch logs and stats in real-time

## 🎮 Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `1` / `l` | Logs view |
| `2` / `s` | Statistics |
| `3` / `p` | Preferences |
| `4` / `i` | Import emails |
| `5` / `e` | Pending emails |
| `B` | Boot/Start sending |
| `a` | Abort/Stop sending |
| `c` | Clear logs |
| `h` | Toggle help |
| `q` | Quit |

## 📊 Database Format

The `data.txt` file uses a simple format:

```
2026-01-03T10:30:00Z ; PENDING ; user@example.com
2026-01-03T10:31:00Z ; DONE ; another@example.com
2026-01-03T10:32:00Z ; FAILED ; failed@example.com ; Error: timeout
```

Status values: `PENDING`, `SENDING`, `DONE`, `FAILED`, `UNSUBSCRIBED`

## 🔧 Advanced Features

### Email Import

Place text files in the working directory and use the Import tab to extract emails automatically:

```
Press 4/i → Select file → Press Enter
```

The app uses regex to find all valid email addresses in the file.

### Template Variables

Use placeholders in your HTML template:

```html
<p>Hello {{email}},</p>
```

### Rate Limiting

Configure delay in Preferences tab or edit `config.yaml`:

```yaml
mail:
  delay_seconds: 30  # Wait 30 seconds between sends
```

## 🏗️ Architecture

```
bulkmail-tui/
├── main.go       # Entry point
├── app.go        # Core business logic
├── tui.go        # Terminal UI
├── types.go      # Data structures
├── database.go   # Data persistence
├── mail.go       # Email sending
├── config.go     # Configuration
└── samples.go    # Sample generators
```

## 🛠️ Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [gomail](https://gopkg.in/gomail.v2) - Email sending
- [fsnotify](https://github.com/fsnotify/fsnotify) - File watching
- [yaml.v3](https://gopkg.in/yaml.v3) - Configuration parsing

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🐛 Known Issues

- None at the moment

## 📮 Support

For issues and questions, please open a GitHub issue.

## 🙏 Acknowledgments

- Inspired by [Listmonk](https://github.com/knadh/listmonk)
- Built with ❤️ using [Charm](https://charm.sh/) tools

---

⭐ **Star this repo if you find it useful!**
