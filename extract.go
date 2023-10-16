package mail

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"

	"go.devnw.com/ds/trees/nary"
	"go.devnw.com/mail/ct"
)

func Extract(e *Email, body io.Reader, tpe, boundary string) error {
	extractors := map[string]SubType{
		ct.MIXED.String(): ct.MIXED,
		ct.ALT.String():   ct.ALT,
		ct.REL.String():   ct.REL,
		ct.DIG.String():   ct.DIG,
		ct.SIGN.String():  ct.SIGN,
		ct.ENC.String():   ct.ENC,
		ct.PLAIN.String(): ct.PLAIN,
		ct.HTML.String():  ct.HTML,
	}

	extractor, ok := extractors[tpe]
	if !ok {
		part, ok := body.(*multipart.Part)
		if !ok || part.FileName() == "" {
			return errors.New("unknown content type")
		}

		// Extract the attachments
		return attachments(e, part)
	}

	return extractor.Extract(e, body, boundary)
}

type Section interface {
	io.Reader
}

func Extract(r io.Reader, boundary string) (nary.Tree[io.Reader], error) {
	out := Tree[*multipart.Part]{}

	root := out.AddRoot(r)

	mr := multipart.NewReader(body, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		ct, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}

	}

	return out, nil
}
