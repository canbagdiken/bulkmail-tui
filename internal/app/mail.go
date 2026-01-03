// mail.go: Email gönderme işlemleri için fonksiyonlar

package app

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

// htmlToPlainText: HTML'i basit plain text'e çevirir
func htmlToPlainText(html string) string {
	// {{email}} gibi placeholder'ları koru
	text := html
	
	// <br>, <br/>, <br /> -> \n
	text = regexp.MustCompile(`(?i)<br\s*/?\s*>`).ReplaceAllString(text, "\n")
	
	// </p>, </div>, </h1>, </h2>, vb -> \n\n
	text = regexp.MustCompile(`(?i)</(?:p|div|h[1-6]|li|tr)>`).ReplaceAllString(text, "\n\n")
	
	// <a href="url">text</a> -> text (url)
	text = regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["'][^>]*>([^<]+)</a>`).ReplaceAllString(text, "$2 ($1)")
	
	// Tüm HTML tag'lerini kaldır
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")
	
	// HTML entity'leri
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	
	// 3+ ardışık \n -> \n\n
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	
	// Boşlukları temizle
	text = strings.TrimSpace(text)
	
	return text
}

// SendMail: Belirtilen config ile email gönderir
func SendMail(cfg *Config, to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(cfg.SMTP.FromEmail, cfg.SMTP.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetHeader("Reply-To", cfg.SMTP.FromEmail)
	m.SetHeader("MIME-Version", "1.0")
	m.SetHeader("Message-ID", fmt.Sprintf("<%d.%s@%s>", time.Now().Unix(), strings.Split(to, "@")[0], cfg.SMTP.Host))
	m.SetHeader("Date", time.Now().Format(time.RFC1123Z))
	m.SetHeader("List-Unsubscribe", fmt.Sprintf("<https://ipieconference.org/unsubscribe/?email=%s>", to))
	m.SetHeader("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")

	// Template'deki {{email}} placeholder'ını değiştir
	body := strings.ReplaceAll(htmlBody, "{{email}}", to)
	
	// HTML'den plain text otomatik üret
	plainText := htmlToPlainText(body)
	
	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", body)

	d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	
	if cfg.SMTP.Port == 465 {
		d.SSL = true
	} else {
		d.TLSConfig = &tls.Config{
			ServerName:         cfg.SMTP.Host,
			InsecureSkipVerify: false,
		}
	}

	return d.DialAndSend(m)
}

