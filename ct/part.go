package ct

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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

	// TODO: add a metadata section for iocs to be added to individual parts

	// TODO: Add hash, and setup so that it's calculated and cached on each
	// readthrough using a TEE
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
	mt, params, err := mime.ParseMediaType(attrs.Get(TYPE.String()))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(mt, MULTIPART.String()) {
		if params[BOUNDARY] != "" {
			body, err = multipart.NewReader(body, params[BOUNDARY]).NextPart()
			if err != nil {
				return nil, err
			}
		}

		// TODO: Add hash as secondary while reading the io.Reader
		// TODO: As the initial read is going through this should scan for links/emails
		// TODO: As the initial read is going through this should parse the HTML and
		// pull all links, emails, contact information, IP addresses, etc... for
		// threat intel feeds
		return &Part{
			mediaType: mt,
			body:      body,
		}, nil
	}

	children, err := Extract(ctx, params, body)
	if err != nil {
		return nil, err
	}

	return &Part{
		mediaType: mt,
		headers:   attrs,
		params:    ToAttributes(params),
		body:      body,
		children:  children,
	}, err
}
