package mail

import (
	"context"
	"io"
	"net/mail"

	"go.devnw.com/mail/ct"
)

type Email struct {
	io.Reader
	*Header

	HTML string
	Text string

	Attachments []Attachment
	Embedded    []Embedded
}

// Attachment with filename, content type and data (as a io.Reader).
type Attachment struct {
	*ct.Part
	FileName    string
	ContentType string
	Data        io.Reader
}

// Embedded with content id, content type and data (as a io.Reader).
type Embedded struct {
	*ct.Part
	CID         string
	ContentType string
	Data        io.Reader
}

func (e *Email) Decode(ctx context.Context, r io.Reader) (err error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}

	e.Header = &Header{}
	err = e.Header.Decode(ctx, msg.Header)
	if err != nil {
		return err
	}

	// Extract the body
	e.Reader = msg.Body

	return nil
}
