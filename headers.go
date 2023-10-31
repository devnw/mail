package mail

import (
	"net/mail"

	"github.com/mnako/letters"
)

type Addresses []*mail.Address

type Headers struct {
	letters.Headers
}
