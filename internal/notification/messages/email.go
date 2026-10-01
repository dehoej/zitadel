package messages

import (
	"fmt"
	"mime"
	"mime/quotedprintable"
	"regexp"
	"strings"
	"time"

	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/notification/channels"
)

var (
	isHTMLRgx = regexp.MustCompile(`.*<html.*>.*`)
	lineBreak = "\r\n"
)

var _ channels.Message = (*Email)(nil)

type Email struct {
	Recipients          []string
	BCC                 []string
	CC                  []string
	SenderEmail         string
	SenderName          string
	ReplyToAddress      string
	Subject             string
	Content             string
	TriggeringEventType eventstore.EventType
}

func (msg *Email) GetContent() (string, error) {
	headers := make(map[string]string)
	from := msg.SenderEmail
	if msg.SenderName != "" {
		from = fmt.Sprintf("%s <%s>", bEncodeWord(msg.SenderName), msg.SenderEmail)
	}
	headers["From"] = from
	if msg.ReplyToAddress != "" {
		headers["Reply-to"] = msg.ReplyToAddress
	}
	headers["To"] = strings.Join(msg.Recipients, ", ")
	headers["Cc"] = strings.Join(msg.CC, ", ")
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s"+lineBreak, k, v)
	}

	//default mime-type is html
	mime := "MIME-Version: 1.0" + lineBreak + "Content-Type: text/html; charset=\"UTF-8\"" + lineBreak
	if !isHTML(msg.Content) {
		mime = "MIME-Version: 1.0" + lineBreak + "Content-Type: text/plain; charset=\"UTF-8\"" + lineBreak
	}
	mime += "Content-Transfer-Encoding: quoted-printable" + lineBreak + lineBreak

	body, err := qpEncode(msg.Content)
	if err != nil {
		return "", err
	}
	subject := "Subject: " + bEncodeWord(msg.Subject) + lineBreak
	message += subject + mime + body

	return message, nil
}

func (msg *Email) GetTriggeringEventType() eventstore.EventType {
	return msg.TriggeringEventType
}

func isHTML(input string) bool {
	return isHTMLRgx.MatchString(input)
}

// returns the content as quoted-printable (RFC 2045) to keep the body 7bit
// and the lines within the limits of RFC 5322
func qpEncode(content string) (string, error) {
	var b strings.Builder
	w := quotedprintable.NewWriter(&b)
	if _, err := w.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}

// returns a RFC1342 "B" encoded string to allow non-ascii characters
func bEncodeWord(word string) string {
	return mime.BEncoding.Encode("UTF-8", word)
}
