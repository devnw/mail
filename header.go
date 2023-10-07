package mail

import (
	"context"
	"errors"
	"mime"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"log/slog"
)

// https://www.w3.org/Protocols/rfc1341/7_1_Text.html
const DEFAULTCT = "text/plain; charset=us-ascii"

type Header struct {
	Subject string `mail:"Subject,mime/word"`

	From    []*mail.Address `mail:"From,mail/addresses"`
	ReplyTo []*mail.Address `mail:"Reply-To,mail/addresses"`
	To      []*mail.Address `mail:"To,mail/addresses"`
	Cc      []*mail.Address `mail:"Cc,mail/addresses"`
	Bcc     []*mail.Address `mail:"Bcc,mail/addresses"`

	ContentType *Media `mail:"Content-Type,mime/media"`

	Date      time.Time `mail:"Date,mail/date"`
	MessageID string    `mail:"Message-ID"`

	ResentFrom      []*mail.Address `mail:"Resent-From,mail/addresses"`
	ResentTo        []*mail.Address `mail:"Resent-To,mail/addresses"`
	ResentCc        []*mail.Address `mail:"Resent-Cc,mail/addresses"`
	ResentBcc       []*mail.Address `mail:"Resent-Bcc,mail/addresses"`
	ResentDate      *time.Time      `mail:"Resent-Date,mail/date"`
	ResentMessageID string

	InReplyTo  []string `mail:"In-Reply-To"`
	References []string `mail:"References"`

	// Security
	AuthenticationResults string   `mail:"Authentication-Results"`
	DkimSignature         string   `mail:"DKIM-Signature"`
	DomainKeySignature    string   `mail:"DomainKey-Signature"`
	ReceivedSPF           string   `mail:"Received-SPF"`
	ReceivedDKIM          string   `mail:"Received-DKIM"`
	ReceivedDomainKey     string   `mail:"Received-DomainKey"`
	Received              []string `mail:"Received"`

	Additional map[string][]string `mail:"-"`
}

func (h *Header) Decode(ctx context.Context, header mail.Header) error {
	d := new(mime.WordDecoder)
	h.Additional = make(map[string][]string)

	val := reflect.ValueOf(h).Elem()
	tpe := val.Type()

	for i := 0; i < val.NumField(); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			field := tpe.Field(i)
			if !field.IsExported() {
				continue
			}

			tag := field.Tag.Get("mail")
			out := val.Field(i)

			// "-" means ignore
			if tag == "-" {
				continue
			}

			if tag == "" {
				value, exists := header[tag]
				if !exists || len(value) == 0 {
					continue
				}

				h.Additional[field.Name] = value
				continue
			}

			fields := strings.Split(tag, ",")
			key := fields[0]
			dec := ""

			if len(fields) > 1 {
				dec = fields[1]
			}

			switch dec {
			case "mime/word":
				s, err := d.Decode(header.Get(key))
				if err != nil {
					out.SetString(header.Get(key))
					continue
				}

				out.SetString(s)
			case "mail/address":
				s := header.Get(key)
				if strings.Trim(s, " \n") == "" {
					continue
				}

				a, err := mail.ParseAddress(s)
				if err != nil {
					return err
				}

				out.Set(reflect.ValueOf(a))
			case "mail/addresses":
				s, err := header.AddressList(key)
				if err != nil {
					if err == mail.ErrHeaderNotPresent {
						continue
					}

					return err
				}

				out.Set(reflect.ValueOf(s))
			case "mail/date":
				v := header.Get(key)
				if v == "" {
					continue
				}

				t, err := mail.ParseDate(v)
				if err != nil {
					return err
				}

				out.Set(reflect.ValueOf(t))
			case "mime/media":
				v := header.Get(key)
				if v == "" {
					continue
				}

				media, params, err := mime.ParseMediaType(v)
				if err != nil {
					return err
				}

				out.Set(reflect.ValueOf(&Media{
					Type:   media,
					Params: params,
				}))
			default:
				values, exists := header[key]
				if !exists || len(values) == 0 {
					continue
				}

				switch out.Kind() {
				case reflect.Slice:
					out.Set(reflect.ValueOf(values))
				case reflect.String:
					out.SetString(strings.Join(values, ", "))
				case reflect.Float32, reflect.Float64:
					v, err := strconv.ParseFloat(values[0], 64)
					if err != nil {
						return err
					}

					out.SetFloat(v)
				case reflect.Int, reflect.Int8, reflect.Int16,
					reflect.Int32, reflect.Int64:
					v, err := strconv.ParseInt(values[0], 10, 64)
					if err != nil {
						return err
					}

					out.SetInt(v)
				default:
					return errors.New("unsupported type")
				}
			}
		}
	}

	return nil
}

type Media struct {
	Type   string
	Params map[string]string
}

// from DS7PR06MB6869.namprd06.prod.outlook.com (2603:10b6:5:2d1::5) by PH8PR06MB9076.namprd06.prod.outlook.com with HTTPS; Wed, 4 Oct 2023 20:47:53 +0000
// from BN8PR03CA0028.namprd03.prod.outlook.com (2603:10b6:408:94::41) by DS7PR06MB6869.namprd06.prod.outlook.com (2603:10b6:5:2d1::5) with Microsoft SMTP Server (version=TLS1_2, cipher=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384) id 15.20.6838.28; Wed, 4 Oct 2023 20:47:49 +0000
//from BN8NAM04FT065.eop-NAM04.prod.protection.outlook.com (2603:10b6:408:94:cafe::c) by BN8PR03CA0028.outlook.office365.com (2603:10b6:408:94::41) with Microsoft SMTP Server (version=TLS1_2, cipher=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384) id 15.20.6863.25 via Frontend Transport; Wed, 4 Oct 2023 20:47:49 +0000
//from outbound-ip137a.ess.barracuda.com (209.222.82.7) by BN8NAM04FT065.mail.protection.outlook.com (10.13.160.195) with Microsoft SMTP Server (version=TLS1_2, cipher=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384) id 15.20.6863.27 via Frontend Transport; Wed, 4 Oct 2023 20:47:48 +0000
//from mta-174-81-118.atlassian.net.sparkpostmail.com (mta-174-81-118.atlassian.net.sparkpostmail.com [192.174.81.118]) by mx-inbound11-215.us-east-2a.ess.aws.cudaops.com (version=TLSv1.2 cipher=ECDHE-RSA-AES256-GCM-SHA384 bits=256 verify=NO); Wed, 04 Oct 2023 20:47:48 +0000
// https://www.pobox.help/hc/en-us/articles/1500000193602-The-elements-of-a-Received-header
// https://www.rfc-editor.org/rfc/rfc1123
// https://www.rfc-editor.org/rfc/rfc822
// https://www.rfc-editor.org/rfc/rfc2076
// https://stackoverflow.com/questions/504136/parsing-email-received-headers
// https://datatracker.ietf.org/doc/html/rfc2821#section-4.4
// https://datatracker.ietf.org/doc/html/rfc821
// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm

type Received []*Transport

func Decode(ctx context.Context, v []string) (Received, error) {
	var r Received

	for _, s := range v {
		t := &Transport{}
		err := t.Decode(ctx, s)
		if err != nil {
			return nil, err
		}

		r = append(r, t)
	}

	return r, nil
}

type Transport struct {
	ID string

	// https://datatracker.ietf.org/doc/html/rfc2821#section-3.8.2
	// The gateway SHOULD indicate the environment and protocol in the "via"
	// clauses of Received field(s) that it supplies.
	Via string
	For Entity

	// https://www.ibm.com/docs/en/zos/2.2.0?topic=sc-helo-command-identify-domain-name-sending-host-smtp
	// https://www.ietf.org/rfc/rfc5321.txt
	Helo string

	From Entity
	By   Entity
	With With

	Date time.Time
}

var ErrInvalidTransport = errors.New("invalid transport")
var ErrIgnoreTransport = errors.New("ignore transport")

var idReg = regexp.MustCompile(`id <?([^\s<>;]{3,})`)

func (t *Transport) Decode(ctx context.Context, s string) (err error) {
	s, err = normalizeReceived(s)
	if err != nil {
		return err
	}

	lastSemi := strings.LastIndex(s, ";")
	if lastSemi != -1 {
		dt := s[lastSemi+1:]
		if dt != "" {
			// Extract date from the end of the string
			t.Date, _ = mail.ParseDate(dt)
			if err != nil {
				slog.WarnContext(
					ctx,
					"failed to parse date",
					slog.String("received", s),
					slog.String("error", err.Error()),
				)
			}
		}

		// Extract the rest of the string
		s = s[:lastSemi]
	}

	// Extract the ID
	matches := idReg.FindStringSubmatch(s)
	if matches != nil && len(matches) > 1 {
		t.ID = matches[1]
	}

	t.Helo = extractHELO(s)

	return nil
}

var fromR = regexp.MustCompile(`(?i)^\(?from `)
var withLocalFor = regexp.MustCompile(`\bwith\s+local\s+for\b`)
var whiteSpaceR = regexp.MustCompile(`\s+`)

func normalizeReceived(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = whiteSpaceR.ReplaceAllString(s, " ")

	if !fromR.MatchString(s) || !frombyReg(s) {
		return "", ErrIgnoreTransport
	}

	s = normalizeKeywords(s, "by", "with", "for", "id", "via")

	// Exclude lines that don't start with "from"
	if withLocalFor.MatchString(s) {
		return "", ErrIgnoreTransport
	}

	return s, nil
}

func normalizeKeywords(s string, keys ...string) string {
	for _, key := range keys {
		index := strings.Index(s, key)
		if index == -1 {
			continue
		}

		if index > 0 {
			after := index + len(key)
			if after < len(s) {
				if s[after] != ' ' {
					// Splice a space after the keyword
					s = s[:index+len(key)] + " " + s[index+len(key):]
				}
			}

			if s[index-1] != ' ' {
				// Splice a space in front of the keyword
				s = s[:index] + " " + s[index:]
			}
		}
	}

	return s
}

func frombyReg(s string) bool {
	mainPattern := `(?i)^from (\S+) by [^\s;]+ ?;`
	mainRe := regexp.MustCompile(mainPattern)

	matches := mainRe.FindStringSubmatch(s)
	if matches != nil && len(matches) > 1 {
		subPattern := `^\[[\d.]+\]$`
		subRe := regexp.MustCompile(subPattern)

		if !subRe.MatchString(matches[1]) {
			return false
		}
	}

	return true
}

type Entity struct {
	Name string
	FQDN string
	IP   net.IP
}

type With struct {
	Name     string
	Metadata map[string]string
}

var heloR = regexp.MustCompile(`(?i)\bhelo=([-A-Za-z0-9.^+_&:=?!@%*$\\\/]+)(?:[^-A-Za-z0-9.^+_&:=?!@%*$\\\/]|$)`)

var ehloR = regexp.MustCompile(`(?i)\b(?:HELO|EHLO) ([-A-Za-z0-9.^+_&:=?!@%*$\\\/]+)(?:[^-A-Za-z0-9.^+_&:=?!@%*$\\\/]|$)`)

func extractHELO(s string) string {
	// Match HELO
	matches1 := heloR.FindStringSubmatch(s)
	if matches1 != nil && len(matches1) > 1 {
		return matches1[1] // Return the captured group from first pattern
	}

	// Match EHLO
	matches2 := ehloR.FindStringSubmatch(s)
	if matches2 != nil && len(matches2) > 1 {
		return matches2[1] // Return the captured group from second pattern
	}

	return ""
}
