package ct

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"
)

type Attributes map[string][]string

func (a Attributes) Get(key string) string {
	return textproto.MIMEHeader(a).Get(key)
}

type Part struct {
	mediaType string
	headers   Attributes
	params    Attributes

	body io.Reader

	children []*Part
}

func (p *Part) Type() string {
	return p.mediaType
}

func (p *Part) String() string {
	return fmt.Sprintf("%s/%s", p.mediaType, p.params[BOUNDARY])
}

func (p *Part) Boundary() string {
	return p.params.Get(BOUNDARY)
}

func (p *Part) MultiPart() bool {
	return len(p.children) > 0
}

func (p *Part) Encoding() string {
	return p.params.Get(TRANSFERENCODING.String())
}

func (p *Part) Text() string {
	return ""
}

// Passthrough the io.Reader interface.
func (p *Part) Read(b []byte) (int, error) {
	return p.body.Read(b)
}

func Parse(
	ctx context.Context,
	attrs Attributes,
	body io.Reader,
) (*Part, error) {
	ct, params, err := mime.ParseMediaType(attrs.Get(TYPE.String()))
	if err != nil {
		return nil, nil, err
	}

	if !strings.HasPrefix(ct, MULTIPART.String()) {
		fmt.Println(ct)
		fmt.Println(attrs.Get(TRANSFERENCODING.String()))
		return nil, nil, fmt.Errorf("invalid media type: %s", ct)
	}

	Extract(ctx, params, body)
}

// Parse takes the full Content-Type header/param value and extracts
// the mime media type information and updates this Part with the parsed
// information for analysis.
func (p *Part) Parse(ctx context.Context) (err error) {
	parts, params, err := Extract(ctx, p.headers, p.body)
	if err != nil {
		return err
	}

	a := Attributes{}
	for k, v := range params {
		a[k] = []string{v}
	}

	p.params = a
	p.children = parts

	return nil
}
