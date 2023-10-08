package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"path"
	"sync"
	"time"

	"go.atomizer.io/stream"
)

func main() {
	// Iterate recursively through all directories and open all .eml
	// files.  For each file, extract the Recevied headers and write them
	// to a file.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fname string
	var ext string

	flag.StringVar(&fname, "o", "recv_headers.txt", "output file")
	flag.StringVar(&ext, "ext", ".eml", "file extension")
	flag.Parse()

	headers := processfile(ctx, readfiles(ctx, filenames(ctx, ".", ext)))

	outfile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case h, ok := <-headers:
			if !ok {
				return
			}

			for _, v := range h {
				if v == "" {
					continue
				}

				fmt.Fprintln(outfile, v)
			}

			fmt.Fprintln(outfile, "") // blank line between messages
		}
	}
}

type file struct {
	path string
	body []byte
}

func processfile(ctx context.Context, files <-chan *file) <-chan []string {
	s := stream.Scaler[*file, []string]{
		Wait: time.Nanosecond,
		Life: time.Millisecond,
		Fn: func(ctx context.Context, in *file) ([]string, bool) {
			msg, err := mail.ReadMessage(bytes.NewReader(in.body))
			if err != nil {
				return nil, false
			}

			return msg.Header["Received"], true
		},
	}

	out, err := s.Exec(ctx, files)
	if err != nil {
		panic(err)
	}

	return out
}

func readfiles(ctx context.Context, files <-chan string) <-chan *file {
	out := make(chan *file)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case path, ok := <-files:
				if !ok {
					return
				}

				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case out <- &file{path, data}:
				}
			}
		}
	}()

	return out
}

func filenames(ctx context.Context, dir, ext string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		files, err := os.ReadDir(dir)
		if err != nil {
			slog.Error(err.Error())
			return
		}

		wg := sync.WaitGroup{}
		for _, file := range files {
			if !file.IsDir() {
				i, err := file.Info()
				if err != nil {
					continue
				}

				fext := path.Ext(i.Name())
				if fext != ext {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case out <- path.Join(dir, i.Name()):
				}

				continue
			}

			i, err := file.Info()
			if err != nil {
				return
			}

			wg.Add(1)
			go func(d os.FileInfo) {
				defer wg.Done()

				stream.Pipe(
					ctx,
					filenames(
						ctx,
						path.Join(dir, d.Name()),
						ext),
					out)
			}(i)
		}

		wg.Wait()
	}()

	return out
}
