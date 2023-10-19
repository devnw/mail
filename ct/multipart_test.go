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
	"github.com/stretchr/testify/assert"
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
			t.Run(path, func(t *testing.T) {
				if d.IsDir() {
					return
				}

				assert.NoError(t, err)

				testeml, err := os.OpenFile(path, os.O_RDONLY, 0644)
				assert.NoError(t, err)

				msg, err := mail.ReadMessage(testeml)
				assert.NoError(t, err)

				p, err := Parse(ctx, HtoA(msg.Header), io.NopCloser(msg.Body))
				assert.NoError(t, err)

				defer p.Close()
			})

			return nil
		})

	if err != nil {
		t.Fatal(err)
	}
}
