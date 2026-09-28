package contact

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Mailer は問い合わせ内容をメールで送信する。SMTP経由の実装（AWS SESのSMTP
// インターフェース等）を想定しているが、featureのコードからは実装詳細を
// 隠蔽する。
type Mailer interface {
	Send(name, email, message string) error
}

type smtpMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	to       string
}

// NewSMTPMailer は環境変数からSMTP接続情報を読み込む。
// CONTACT_SMTP_HOST / CONTACT_SMTP_PORT / CONTACT_SMTP_USERNAME /
// CONTACT_SMTP_PASSWORD / CONTACT_FROM_EMAIL / CONTACT_TO_EMAIL のいずれかが
// 未設定の場合、Send は常にエラーを返す（未設定＝問い合わせ機能が無効、という
// admin.ADMIN_PASSWORD と同じ扱い）。
func NewSMTPMailer() Mailer {
	return &smtpMailer{
		host:     os.Getenv("CONTACT_SMTP_HOST"),
		port:     os.Getenv("CONTACT_SMTP_PORT"),
		username: os.Getenv("CONTACT_SMTP_USERNAME"),
		password: os.Getenv("CONTACT_SMTP_PASSWORD"),
		from:     os.Getenv("CONTACT_FROM_EMAIL"),
		to:       os.Getenv("CONTACT_TO_EMAIL"),
	}
}

func (m *smtpMailer) configured() bool {
	return m.host != "" && m.port != "" && m.username != "" && m.password != "" && m.from != "" && m.to != ""
}

func (m *smtpMailer) Send(name, email, message string) error {
	if !m.configured() {
		return fmt.Errorf("contact: CONTACT_SMTP_* / CONTACT_FROM_EMAIL / CONTACT_TO_EMAIL is not fully configured")
	}

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	body := buildMessage(m.from, m.to, email, name, message)

	return smtp.SendMail(m.host+":"+m.port, auth, m.from, []string{m.to}, []byte(body))
}

// buildMessage はRFC 5322準拠の生メール（ヘッダー+本文）を組み立てる。
// SESのSMTPインターフェースは送信元(From)がSESで検証済みのアドレスである
// ことを要求するため、返信はReply-Toで問い合わせ者本人のアドレスへ向ける。
func buildMessage(from, to, replyTo, name, message string) string {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Reply-To: " + replyTo,
		"Subject: " + encodeSubject(fmt.Sprintf("[Code Beaver] %sさんからお問い合わせ", name)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Date: " + time.Now().Format(time.RFC1123Z),
	}
	body := fmt.Sprintf("お名前: %s\nメールアドレス: %s\n\n%s\n", name, replyTo, message)
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

// encodeSubject はSubjectヘッダーの非ASCII文字をRFC 2047のencoded-wordに変換する。
func encodeSubject(subject string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
}
