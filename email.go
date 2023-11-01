package mail

import (
	"context"
	"io"
	"net/url"

	"github.com/mnako/letters"
	"go.atomizer.io/stream"
)

func Load(r io.ReadCloser) (*Email, error) {
	defer r.Close()

	email, err := letters.ParseEmail(r)
	if err != nil {
		return nil, err
	}

	return &Email{
		email,
		&Headers{email.Headers},
	}, nil
}

// Email is a wrapper around the letters.Email type which provides
// additional functionality for extracting links and hashes from
// the email as well as other useful information such as the ability
// to evaluate headers for the email, receipt tracking, DMARC, SPF, and DKIM.
type Email struct {
	letters.Email
	h *Headers
}

func (e *Email) Metrics(ctx context.Context) (*Metrics, error) {
	return compile(ctx, e)
}

// Hashes returns a slice of hashes for the email including the
// hashes of all attachments and inline files.
func (e *Email) Hashes(ctx context.Context) ([][]byte, error) {
	seen := make(map[string]struct{})
	hashes := stream.Intercept(ctx, stream.FanIn(
		ctx,
		hash(ctx, attachs(ctx, e.AttachedFiles)),
		hash(ctx, embeds(ctx, e.InlineFiles)),
	), func(ctx context.Context, data []byte) ([]byte, bool) {
		_, ok := seen[string(data)]
		if ok {
			return nil, false // filter dups
		}

		return data, true
	})

	var out [][]byte
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case h, ok := <-hashes:
			if !ok {
				return out, nil
			}

			out = append(out, h)
		}
	}
}

// Links returns a slice of all links in the email from both the
// text and html bodies, as well as all attachments and inline files
// if they are supported file types
//
// If the links are safe links, they will be stripped of the
// safe link wrapper and the original link will be returned.
func (e *Email) Links() ([]*url.URL, error) {
	// Evaluate the text body for links
	out := urls(e.Text)

	// TODO: strip <a href""> tags from HTML

	// Evaluate the html body non href links for urls
	out = append(out, urls(e.HTML)...)

	// Evaluate the attachments for links

	// Evaluate the inline files for links

	return out, nil
}

func (e *Email) Headers() *Headers {
	return e.h
}
