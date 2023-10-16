package ct

import (
	"context"
	"fmt"
	"io"
	"net/textproto"
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

// Passthrough the io.Reader interface
func (p *Part) Read(b []byte) (int, error) {
	return p.body.Read(b)
}

// Parse takes the full Content-Type header/param value and extracts
// the mime media type information and updates this Part with the parsed
// information for analysis
func (p *Part) Parse(ctx context.Context) (err error) {
	parts, err := Extract(ctx, p.params, p.body)
	if err != nil {
		return err
	}

	p.children = parts

	return nil
}
