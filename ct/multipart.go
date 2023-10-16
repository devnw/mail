package ct

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
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

func Extract[T ~map[string][]string](
	ctx context.Context,
	attrs T,
	body io.Reader,
) (Multipart, error) {
	headers := Attributes(attrs)

	fmt.Println(headers)

	m := Multipart{}
	ct, params, err := mime.ParseMediaType(headers.Get(TYPE.String()))
	if err != nil {
		return nil, err
	}

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
			if err != nil {
				if err == io.EOF {
					return m, nil
				}

				return nil, err
			}

			p := &Part{
				mediaType: ct,
				body:      pt,
				headers:   Attributes(pt.Header),
			}

			err = p.Parse(ctx)
			if err != nil {
				return nil, err
			}

			m = append(m, p)
		}
	}

	return m, nil
}

func ToAttributes[T map[string]string](p T) Attributes {
	a := Attributes{}
	for k, v := range p {
		a[k] = []string{v}
	}

	return a
}
