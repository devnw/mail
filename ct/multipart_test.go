package ct

import (
	"context"
	"io"
	"io/fs"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func init() {

	spew.Config.DisableMethods = true
}

func Test_Multipart_Extract(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := filepath.WalkDir(
		"../testdata/phishing_corpus/1/",
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if err != nil {
				return err
			}

			testeml, err := os.OpenFile(path, os.O_RDONLY, 0644)
			if err != nil {
				return err
			}

			msg, err := mail.ReadMessage(testeml)
			if err != nil {
				return err
			}

			p, err := Parse(ctx, HtoA(msg.Header), io.NopCloser(msg.Body))
			if err != nil {
				return err
			}

			defer p.Close()

			return nil
		})

	if err != nil {
		t.Fatal(err)
	}
}
