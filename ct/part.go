package ct

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"

	"errors"
)

type Attributes map[string][]string

func (a Attributes) Get(key string) string {
	return textproto.MIMEHeader(a).Get(key)
}

type Part struct {
	mediaType string
	headers   Attributes
	params    Attributes

	body io.ReadCloser

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
	return p.headers.Get(TRANSFERENCODING.String())
}

func (p *Part) Text() string {
	return ""
}

// Passthrough the io.Reader interface.
func (p *Part) Read(b []byte) (int, error) {
	return p.body.Read(b)
}

func (p *Part) Close() error {
	for _, c := range p.children {
		_ = c.Close()
	}

	return p.body.Close()
}

func Parse(
	ctx context.Context,
	attrs Attributes,
	body io.ReadCloser,
) (*Part, error) {
	mt, params, err := mime.ParseMediaType(
		normalizeMediaType(attrs.Get(TYPE.String())),
	)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf(
			"failed to parse media type %s",
			attrs.Get(TYPE.String()),
		))
	}

	if !strings.HasPrefix(mt, MULTIPART.String()) {
		//if params[BOUNDARY] != "" {
		//	body, err = multipart.NewReader(body, params[BOUNDARY]).NextPart()
		//	if err == io.EOF {
		//		return nil, nil
		//	}

		//	if err != nil {
		//		return nil, err
		//	}
		//}

		// TODO: Add hash as secondary while reading the io.Reader
		// TODO: As the initial read is going through this should scan for links/emails
		// TODO: As the initial read is going through this should parse the HTML and
		// pull all links, emails, contact information, IP addresses, etc... for
		// threat intel feeds
		p := &Part{
			mediaType: mt,
			headers:   attrs,
			params:    ToAttributes(params),
			body:      body,
		}

		fmt.Printf(
			"multipart: %v; type: %s; encoding: %s; boundary: %s;\n",
			p.MultiPart(),
			p.Type(),
			p.Encoding(),
			p.Boundary(),
		)

		return p, nil
	}

	children, err := Extract(ctx, params, body)
	if err != nil {
		return nil, err
	}

	p := &Part{
		mediaType: mt,
		headers:   attrs,
		params:    ToAttributes(params),
		body:      body,
		children:  children,
	}

	fmt.Printf(
		"multipart: %v; type: %s; encoding: %s; boundary: %s;\n",
		p.MultiPart(),
		p.Type(),
		p.Encoding(),
		p.Boundary(),
	)

	return p, err
}

func normalizeMediaType(mt string) string {
	// Add a space after each ; where one doesn't exist
	for i, r := range mt {
		if r == ';' && i+1 < len(mt) && mt[i+1] != ' ' {
			mt = fmt.Sprintf("%s %s", mt[:i+1], mt[i+1:])
		}
	}

	out := strings.ToLower(strings.TrimSpace(mt))

	fmt.Printf("normalized media type: %s\n", out)

	return out
}
