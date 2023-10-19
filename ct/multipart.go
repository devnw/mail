package ct

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
)

type CONTENT string

func (c CONTENT) String() string { return string(c) }

const (
	TYPE CONTENT = "Content-Type"

	// If a MIME part is 7bit, the Content-Transfer-Encoding header is
	// optional. MIME parts with any other transfer encoding must contain
	// a Content-Transfer-Encoding header. If the MIME part is a multipart
	// content type, the part should not have an encoding of base64 or
	// quoted-printable.
	TRANSFERENCODING CONTENT = "Content-Transfer-Encoding"
	DISPOSITION      CONTENT = "Content-Disposition"
	BOUNDARY                 = "boundary"
)

var ErrMissingBoundary = errors.New("missing boundary")

type Multipart []*Part

func (m Multipart) Type() string { return "multipart" }

func Extract(
	ctx context.Context,
	params map[string]string,
	body io.Reader,
) (Multipart, error) {
	var m Multipart

	boundary := params[BOUNDARY]
	if boundary == "" {
		return nil, ErrMissingBoundary
	}

	mr := multipart.NewReader(body, boundary)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			pt, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				return m, nil
			}

			if err != nil {
				return nil, err
			}

			p, err := Parse(ctx, HtoA(pt.Header), pt)
			if err != nil {
				return nil, err
			}

			m = append(m, p)
		}
	}

	return m, nil
}

func HtoA[T ~map[string][]string](h T) Attributes {
	attr := Attributes{}
	for k, v := range h {
		attr[k] = v
	}

	return attr
}

func AtoP(a Attributes) map[string]string {
	out := map[string]string{}

	for k, _ := range a {
		out[k] = a.Get(k)
	}

	return out
}

func ToAttributes[T map[string]string](p T) Attributes {
	a := Attributes{}
	for k, v := range p {
		a[k] = []string{v}
	}

	return a
}
