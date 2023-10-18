package ct

import (
	"context"
	"net/mail"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_Multipart_Extract(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testeml, err := os.OpenFile("./testdata/multipart/test.eml", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := mail.ReadMessage(testeml)
	if err != nil {
		t.Fatal(err)
	}

	p, err := Parse(ctx, HtoA(msg.Header), msg.Body)
	if err != nil {
		t.Fatal(err)
	}

	spew.Config.DisableMethods = true
	spew.Dump(p)
}
