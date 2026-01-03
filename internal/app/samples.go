package app

import (
	"os"
	"time"
)

func CreateSampleConfig() error {
	sample := `smtp:
  host: smtp.example.com
  port: 587
  username: yourusername
  password: yourpassword
  from_email: your@email.com
  from_name: Your Name

mail:
  delay_seconds: 30
  subject: "Sample Subject"
  template: mail.html

database:
  path: data.txt
`
	return os.WriteFile("config.yaml", []byte(sample), 0644)
}

func CreateSampleTemplate(path string) error {
	sample := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Sample Email</title>
</head>
<body>
<h1>Sample Email</h1>
<p>Hello {{email}},</p>
<p>This is a sample email template.</p>
<p>Customize this template as needed.</p>
</body>
</html>`
	return os.WriteFile(path, []byte(sample), 0644)
}

func CreateSampleData(path string) error {
	sample := time.Now().Format(time.RFC3339) + " ; " + StatusPending + " ; sample@example.com\n"
	return os.WriteFile(path, []byte(sample), 0644)
}

