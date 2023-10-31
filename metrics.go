package mail

import (
	"context"
	"net/mail"
	"net/url"
	"time"
)

type MessageID string

type Metrics struct {
	Send

	Subject string

	Comment  string
	Keywords []string

	Resent *Send

	ReplyTo []*mail.Address

	InReplyTo  []MessageID
	References []MessageID

	// DKIM  bool
	// SPF   bool
	// DMARC bool

	Attachments int
	Embeds      int

	Hashes [][]byte
	Links  []*url.URL
}

func (m *Metrics) Date() time.Time {
	return m.Send.Date
}

type Send struct {
	ID MessageID

	Date time.Time

	Sender *mail.Address
	From   []*mail.Address

	To  []*mail.Address
	Cc  []*mail.Address
	Bcc []*mail.Address
}

func compile(ctx context.Context, e *Email) (*Metrics, error) {
	m := &Metrics{
		Send: Send{
			ID:     MessageID(e.h.MessageID),
			Date:   e.h.Date,
			To:     e.h.To,
			Sender: e.h.Sender,
			From:   e.h.From,
			Cc:     e.h.Cc,
			Bcc:    e.h.Bcc,
		},

		//DKIM:  e.DKIM,
		//SPF:   e.SPF,
		//DMARC: e.DMARC,

		Attachments: len(e.AttachedFiles),
		Embeds:      len(e.InlineFiles),
	}

	for _, r := range e.h.InReplyTo {
		m.InReplyTo = append(m.InReplyTo, MessageID(r))
	}

	if len(e.h.ResentFrom) > 0 || !e.h.ResentDate.IsZero() {
		m.Resent = &Send{
			ID:     MessageID(e.h.ResentMessageID),
			Date:   e.h.ResentDate,
			To:     e.h.ResentTo,
			Sender: e.h.ResentSender,
			From:   e.h.ResentFrom,
			Cc:     e.h.ResentCc,
			Bcc:    e.h.ResentBcc,
		}
	}

	hashes, err := e.Hashes(ctx)
	if err != nil {
		return m, err
	}

	m.Hashes = hashes

	links, err := e.Links()
	if err != nil {
		return m, err
	}

	m.Links = links

	return m, nil
}
