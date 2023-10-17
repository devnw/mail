package ct

import (
	"context"
	"net/mail"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_Multipart_Extract(t *testing.T) {
	ncceml, err := os.OpenFile("./testdata/multipart/ncc.eml", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := mail.ReadMessage(ncceml)
	if err != nil {
		t.Fatal(err)
	}

	m, p, err := Extract(context.Background(), Attributes(msg.Header), msg.Body)
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(p)
	spew.Dump(m)
}
