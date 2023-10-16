package ct

import (
	"errors"
	"fmt"
	"io"
)

type Extractor func(io.Reader, string) ([]*Part, error)

/*
RFC 1426, SMTP Service Extension for 8bit-MIMEtransport. J. Klensin, N. Freed, M. Rose, E. Stefferud, D. Crocker. February 1993.
RFC 1847, Security Multiparts for MIME: Multipart/Signed and Multipart/Encrypted
RFC 3156, MIME Security with OpenPGP
RFC 2045, MIME Part One: Format of Internet Message Bodies
RFC 2046, MIME Part Two: Media Types. N. Freed, Nathaniel Borenstein. November 1996.
RFC 2047, MIME Part Three: Message Header Extensions for Non-ASCII Text. Keith Moore. November 1996.
RFC 4288, MIME Part Four: Media Type Specifications and Registration Procedures.
RFC 4289, MIME Part Four: Registration Procedures. N. Freed, J. Klensin. December 2005.
RFC 2049, MIME Part Five: Conformance Criteria and Examples. N. Freed, N. Borenstein. November 1996.
RFC 2183, Communicating Presentation Information in Internet Messages: The Content-Disposition Header Field. Troost, R., Dorner, S. and K. Moore. August 1997.
RFC 2231, MIME Parameter Value and Encoded Word Extensions: Character Sets, Languages, and Continuations. N. Freed, K. Moore. November 1997.
RFC 2387, The MIME Multipart/Related Content-type
RFC 1521, Mechanisms for Specifying and Describing the Format of Internet Message Bodies
RFC 7578, Returning Values from Forms: multipart/form-data
*/

// https://www.iana.org/assignments/media-types/media-types.xhtml
type Type string

type SubType struct {
	Type    Type
	Name    string
	Extract Extractor
}

func (s SubType) String() string {
	return fmt.Sprintf("%s/%s", s.Type, s.Name)
}

const (
	// https://en.wikipedia.org/wiki/MIME#Multipart_subtypes
	MULTIPART   Type = "multipart"
	TEXT        Type = "text"
	IMAGE       Type = "image"
	AUDIO       Type = "audio"
	VIDEO       Type = "video"
	APPLICATION Type = "application"
	MESSAGE     Type = "message"
)

//func Extract(body io.Reader, ct, boundary string) error {
//	t, err := GetExtractor(ct)
//	if err != nil {
//		return err
//	}
//
//	parts, err := t.Extract(body, boundary)
//	if err != nil {
//		return err
//	}
//
//	for _, part := range parts {
//		// Extract Parts
//		err = Extract(part.Body, part.MediaType, part.Params[BOUNDARY])
//		if err != nil {
//			return err
//		}
//
//		// Extract Attachments
//	}
//
//	return nil
//}
//
//func GetExtractor(ct string) (SubType, error) {
//	return SubType{}, nil
//}

//var (
//	ALT = SubType{MULTIPART, "alternative", multiAlt}
//	REL = SubType{MULTIPART, "related", multiRel}
//
//	// https://en.wikipedia.org/wiki/MIME#digest
//	DIG = SubType{MULTIPART, "digest", noop}
//
//	// https://www.oreilly.com/library/view/programming-internet-email/9780596802585/ch05s03s01.html
//	SIGN = SubType{MULTIPART, "signed", multiSign}
//
//	// https://www.iana.org/assignments/media-types/multipart/encrypted#:~:text=The%20multipart%2Fencrypted%20content%20type,value%20of%20the%20protocol%20parameter.
//	// https://www.oreilly.com/library/view/programming-internet-email/9780596802585/ch05s03s02.html
//	ENC = SubType{MULTIPART, "encrypted", multiEnc}
//
//	// https://www.iana.org/assignments/media-types/media-types.xhtml#text
//	PLAIN = SubType{TEXT, "plain", txtPlain}
//	HTML  = SubType{TEXT, "html", txtHTML}
//)
//
//var MIXED = SubType{
//	MULTIPART,
//	"mixed",
//	multiMixed,
//}

var ErrMissingContentType = errors.New("Missing Content-Type Header")

//func NewPart(params Parameters, body io.Reader) (*Part, error) {
//	p := &Part{
//		Params: params,
//	}
//
//	ct, ok := p.Params[CONTENTTYPE]
//	if !ok {
//		return nil, ErrMissingContentType
//	}
//
//	err := p.Parse(ctx, ct)
//	if err != nil {
//		return nil, err
//	}
//
//	p.Body = body
//
//	return p, nil
//}

//func multiAlt(body io.Reader, boundary string) error {
//	return nil
//}
//
//func multiRel(body io.Reader, boundary string) error {
//	return nil
//}
//
//func multiSign(body io.Reader, boundary string) error {
//	return nil
//}
//
//func multiEnc(body io.Reader, boundary string) error {
//	return nil
//}
//
//func txtPlain(body io.Reader, boundary string) error {
//	content, err := io.ReadAll(body)
//	if err != nil {
//		return err
//	}
//
//	e.Text += strings.TrimSuffix(string(content), "\n")
//	return nil
//}
//
//func txtHTML(body io.Reader, boundary string) error {
//	content, err := io.ReadAll(body)
//	if err != nil {
//		return err
//	}
//
//	e.Text += strings.TrimSuffix(string(content), "\n")
//	return nil
//}
//
//func noop(body io.Reader, boundary string) error {
//	slog.Warn("noop extractor executed", slog.String("boundary", boundary))
//	return nil
//}
//
//func attachments(e *Email, part *multipart.Part) error {
//	return nil
//}
