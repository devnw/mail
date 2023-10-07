package mail

import (
	"context"
	"io"
	"net/mail"
)

type Email struct {
	*Header
	body io.Reader

	HTMLBody string
	TextBody string

	// Attachments   []Attachment
	// EmbeddedFiles []EmbeddedFile
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

	e.body = msg.Body

	return nil
}
