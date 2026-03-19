package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"log/slog"

	"github.com/mnako/letters"
)

func attachs(ctx context.Context, a []letters.AttachedFile) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)

		for _, a := range a {
			select {
			case <-ctx.Done():
			case out <- a.Data:
			}
		}
	}()

	return out
}

func embeds(ctx context.Context, e []letters.InlineFile) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)

		for _, e := range e {
			select {
			case <-ctx.Done():
			case out <- e.Data:
			}
		}
	}()

	return out
}

func hash(ctx context.Context, data <-chan []byte) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-data:
				if !ok {
					return
				}

				h := sha256.New()
				_, err := io.Copy(h, bytes.NewReader(d))
				if err != nil {
					slog.WarnContext(
						ctx,
						"Error hashing data",
						"error", err.Error(),
					)
					continue
				}

				out <- h.Sum(nil)
			}
		}
	}()

	return out
}
