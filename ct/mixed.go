package ct

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
)

type mixed struct{}

func (m mixed) String() { return fmt.Sprintf("%s/%s", MULTIPART, "mixed") }

func (m mixed) Extract(params map[string][]string, body io.Reader) ([]*Part, error) {
	ct, params, err := mime.ParseMediaType(part.Header.Get(CONTENTTYPE))
	if err != nil {
		return err
	}

	parts := []*Part{}
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				return parts, nil
			}

			return nil, err
		}

		/*
			Content-Type: application/pdf; name="<filename>.pdf"
			Content-Description: <filename>.pdf
			Content-Disposition: attachment;
			filename="<filename>.pdf"; size=967271;
			creation-date="Mon, 02 Oct 2023 09:12:06 GMT";
			modification-date="Mon, 02 Oct 2023 09:15:01 GMT"
			Content-ID: <46B033A598410F4BB1C4AA0E6C12FD96@namprd06.prod.outlook.com>
			Content-Transfer-Encoding: base64
		*/

		p := &Part{
			Body:   part,
			Header: part.Header,
		}

		p.Parse()

	}

	return parts, nil
}
