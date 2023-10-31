package mail

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeebo/assert"
)

func Test_Email_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loadErrors := 0
	total := 0
	attachments := 0
	embeds := 0

	err := filepath.WalkDir(
		"./testdata/phishing_corpus/1/",
		func(path string, d fs.DirEntry, err error) error {
			t.Run(path, func(t *testing.T) {
				if d.IsDir() {
					return
				}

				assert.NoError(t, err)

				total++

				testeml, err := os.OpenFile(path, os.O_RDONLY, 0644)
				assert.NoError(t, err)

				eml, err := Load(testeml)
				if err != nil {
					t.Logf("error loading email [%s]: %s", path, err)
					loadErrors++
					return
				}

				metrics, err := eml.Metrics(ctx)
				assert.NoError(t, err)

				attachments += metrics.Attachments
				embeds += metrics.Embeds
			})

			return nil
		})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Load Errors: %d", loadErrors)
	t.Logf("Total Emails: %d", total)
	t.Logf("Total Attachments: %d", attachments)
	t.Logf("Total Embeds: %d", embeds)
	t.Logf("Success Rate: %f%%", 100*(float64(total-loadErrors)/float64(total)))
}
