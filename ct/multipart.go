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
	TYPE             CONTENT = "Content-Type"
	TRANSFERENCODING CONTENT = "Content-Transfer-Encoding"
	DISPOSITION      CONTENT = "Content-Disposition"
	BOUNDARY                 = "boundary"
)

var ErrMissingBoundary = errors.New("missing boundary")

type Multipart []*Part

func (m Multipart) Type() string { return "multipart" }

func Extract(
	ctx context.Context,
	attrs Attributes,
	body io.Reader,
) (Multipart, map[string]string, error) {
	m := Multipart{}

	boundary := params[BOUNDARY]
	if boundary == "" {
		return nil, nil, ErrMissingBoundary
	}

	mr := multipart.NewReader(body, boundary)
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()

		default:
			pt, err := mr.NextPart()
			if err != nil {
				if err == io.EOF {
					return m, params, nil
				}

				return nil, nil, err
			}

			p := &Part{
				mediaType: ct,
				body:      pt,
				headers:   Attributes(pt.Header),
			}

			err = p.Parse(ctx)
			if err != nil {
				return nil, nil, err
			}

			m = append(m, p)
		}
	}

	return m, params, nil
}

func ToAttributes[T map[string]string](p T) Attributes {
	a := Attributes{}
	for k, v := range p {
		a[k] = []string{v}
	}

	return a
}
