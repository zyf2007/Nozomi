package server

import (
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

func parseMail(from string, to []string, raw []byte) MailInput {
	input := MailInput{From: from, To: to, Raw: string(raw), Headers: map[string]string{}}
	msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		input.Text = string(raw)
		return input
	}
	for k, v := range msg.Header {
		input.Headers[k] = strings.Join(v, "\n")
	}
	input.Subject = msg.Header.Get("Subject")
	if decoded, err := (&mime.WordDecoder{}).DecodeHeader(input.Subject); err == nil {
		input.Subject = decoded
	}
	mediaType, params, _ := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			b, _ := io.ReadAll(part)
			ct := part.Header.Get("Content-Type")
			if strings.Contains(ct, "text/html") && input.HTML == "" {
				input.HTML = string(b)
			}
			if strings.Contains(ct, "text/plain") && input.Text == "" {
				input.Text = string(b)
			}
		}
	} else {
		b, _ := io.ReadAll(msg.Body)
		if strings.Contains(mediaType, "html") {
			input.HTML = string(b)
		} else {
			input.Text = string(b)
		}
	}
	return input
}
