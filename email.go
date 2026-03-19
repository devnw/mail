package mail

import (
	"context"
	"io"
	"net/url"
	"sync"

	"github.com/mnako/letters"
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
	merged := fanIn(ctx,
		hash(ctx, attachs(ctx, e.AttachedFiles)),
		hash(ctx, embeds(ctx, e.InlineFiles)),
	)

	var out [][]byte
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case h, ok := <-merged:
			if !ok {
				return out, nil
			}

			key := string(h)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}

			out = append(out, h)
		}
	}
}

// fanIn merges multiple channels into a single channel.
func fanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- v:
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
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
