package ct

import "io"

func Decode(body io.Reader, encoding string) ([]byte, error) {

	return nil, nil
}

type Options func(e Encoder) error

type Encoder interface {
	Encode([]byte, ...Options)
}

type Decoder interface {
	Decode(io.Reader) ([]byte, error)
}
