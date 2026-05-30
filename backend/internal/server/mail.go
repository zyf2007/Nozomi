package server

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
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
	parseMailBody(&input, msg.Header.Get("Content-Type"), msg.Body)
	return input
}

func parseMailBody(input *MailInput, contentType string, body io.Reader) {
	mediaType, params, _ := mime.ParseMediaType(contentType)
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			b, _ := io.ReadAll(body)
			parseSinglePart(input, mediaType, nil, b)
			return
		}
		mr := multipart.NewReader(body, boundary)
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			b, _ := io.ReadAll(part)
			parseSinglePart(input, part.Header.Get("Content-Type"), textproto.MIMEHeader(part.Header), b)
		}
		return
	}
	b, _ := io.ReadAll(body)
	parseSinglePart(input, mediaType, nil, b)
}

func parseSinglePart(input *MailInput, contentType string, header textproto.MIMEHeader, raw []byte) {
	mediaType, params, _ := mime.ParseMediaType(contentType)
	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		parseMailBody(input, contentType, strings.NewReader(string(raw)))
	case strings.Contains(mediaType, "text/html") && input.HTML == "":
		input.HTML = decodeBodyText(raw, params["charset"])
	case strings.Contains(mediaType, "text/plain") && input.Text == "":
		input.Text = decodeBodyText(raw, params["charset"])
	default:
		if isAttachmentPart(contentType, header) {
			input.Attachments = append(input.Attachments, parseAttachmentPart(header, raw))
		}
	}
}

func isAttachmentPart(contentType string, header textproto.MIMEHeader) bool {
	disposition := strings.ToLower(header.Get("Content-Disposition"))
	if strings.Contains(disposition, "attachment") {
		return true
	}
	if strings.Contains(disposition, "inline") && header.Get("Content-ID") != "" {
		return true
	}
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType != "" && !strings.HasPrefix(mediaType, "text/") && !strings.HasPrefix(mediaType, "multipart/")
}

func parseAttachmentPart(header textproto.MIMEHeader, raw []byte) MailAttachment {
	contentType := header.Get("Content-Type")
	mediaType, params, _ := mime.ParseMediaType(contentType)
	filename := attachmentFilename(header, params)
	if filename == "" {
		filename = "attachment"
	}
	return MailAttachment{
		Filename:      filename,
		ContentType:   firstNonEmpty(mediaType, "application/octet-stream"),
		ContentBase64: base64.StdEncoding.EncodeToString(raw),
		ContentID:     strings.Trim(header.Get("Content-ID"), "<>"),
		Inline:        strings.Contains(strings.ToLower(header.Get("Content-Disposition")), "inline"),
	}
}

func attachmentFilename(header textproto.MIMEHeader, params map[string]string) string {
	if filename := params["name"]; filename != "" {
		return filename
	}
	if filename := header.Get("Content-Disposition"); filename != "" {
		if _, attrs, err := mime.ParseMediaType(filename); err == nil {
			if v := attrs["filename"]; v != "" {
				return v
			}
		}
	}
	if filename := header.Get("Content-Type"); filename != "" {
		if _, attrs, err := mime.ParseMediaType(filename); err == nil {
			if v := attrs["name"]; v != "" {
				return v
			}
		}
	}
	return ""
}

func decodeBodyText(raw []byte, charset string) string {
	if strings.EqualFold(strings.TrimSpace(charset), "utf-8") || charset == "" {
		return string(raw)
	}
	return string(raw)
}
