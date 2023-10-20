package ct

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"
	"sync"

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

	hash       []byte
	bodyBuffMu sync.Mutex
	bodyBuff   *bytes.Buffer

	body io.ReadCloser

	children []*Part

	// TODO: add a metadata section for iocs to be added to individual
	// parts

	// TODO: Add hash, and setup so that it's calculated and cached on
	// each readthrough using a TEE
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

func (p *Part) Children() []*Part {
	return p.children
}

func (p *Part) Close() (err error) {
	for _, c := range p.children {
		err = errors.Join(err, c.Close())
	}

	if p.body == nil {
		return err
	}

	return errors.Join(err, p.body.Close())
}

var ErrNilBody = errors.New("nil body")

func (p *Part) Hash() ([]byte, error) {
	p.bodyBuffMu.Lock()
	defer p.bodyBuffMu.Unlock()

	if p.body == nil {
		return nil, ErrNilBody
	}

	if p.hash != nil {
		return p.hash, nil
	}

	p.bodyBuff = new(bytes.Buffer)
	sha256 := sha256.New()

	_, err := p.bodyBuff.ReadFrom(io.TeeReader(p.body, sha256))
	if err != nil {
		return nil, err
	}

	return sha256.Sum(nil), nil
}

func Hashes(parts ...*Part) ([][]byte, error) {
	hashes := [][]byte{}

	for _, p := range parts {
		h, err := p.Hash()
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, h)

		ch, err := Hashes(p.Children()...)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, ch...)
	}

	return hashes, nil
}

func Parse(
	ctx context.Context,
	attrs Attributes,
	body io.ReadCloser,
) (*Part, error) {
	normMT := normalizeMediaType(attrs.Get(TYPE.String()))
	mt, params, err := mime.ParseMediaType(normMT)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf(
			"failed to parse media type %s",
			normMT,
		))
	}

	if !strings.HasPrefix(mt, MULTIPART.String()) {
		// if params[BOUNDARY] != "" {
		//	body, err = multipart.NewReader(body, params[BOUNDARY]).NextPart()
		//	if err == io.EOF {
		//		return nil, nil
		//	}

		//	if err != nil {
		//		return nil, err
		//	}
		//}

		// TODO: Add hash as secondary while reading the io.Reader

		// TODO: As the initial read is going through this should scan
		// for links/emails

		// TODO: As the initial read is going through this should parse
		// the HTML and pull all links, emails, contact information,
		// IP addresses, etc... for threat intel feeds
		p := &Part{
			mediaType: mt,
			headers:   attrs,
			params:    ToAttributes(params),
			body:      body,
		}

		fmt.Printf(
			"multipart: %v; type: %s; encoding: %s; boundary: %s;\n",
			len(p.Children()) > 0,
			p.Type(),
			p.Encoding(),
			p.Boundary(),
		)

		return p, nil
	}

	buff := new(bytes.Buffer)

	p := &Part{
		mediaType: mt,
		headers:   attrs,
		params:    ToAttributes(params),
		body:      io.NopCloser(buff),
	}

	p.children, err = Extract(ctx, params, io.TeeReader(body, buff))
	if err != nil {
		return nil, err
	}

	fmt.Printf(
		"multipart: %v; type: %s; encoding: %s; boundary: %s;\n",
		len(p.Children()) > 0,
		p.Type(),
		p.Encoding(),
		p.Boundary(),
	)

	return p, err
}

func normalizeMediaType(mt string) string {
	mt = strings.ReplaceAll(mt, `charset="charset="`, `charset="`)

	// Add a space after each ; where one doesn't exist
	for i, r := range mt {
		if r == ';' && i+1 < len(mt) && mt[i+1] != ' ' {
			mt = fmt.Sprintf("%s %s", mt[:i+1], mt[i+1:])
		}
	}

	mt = strings.ReplaceAll(mt, ` iso-8859-1`, `charset=iso-8859-1`)

	out := strings.ToLower(strings.TrimSpace(mt))

	return out
}
