package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/brigadecore/brigade/v2/internal/rand"
)

var seededRand = rand.NewSeeded()

func ShortSHA(salt, input string) string {
	if salt != "" {
		input = fmt.Sprintf("%s:%s", salt, input)
	}
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)[0:54]
}

// TODO: These aren't technically guaranteed unique, although a collision would
// be extraordinary. There's probably a reliable way of generating tokens that
// are guaranteed unique. We should identify that and transition.
func NewToken(tokenLength int) string {
	const (
		tokenChars = "abcdefghijklmnopqrstuvwxyz" +
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"0123456789"
	)
	b := make([]byte, tokenLength)
	for i := 0; i < tokenLength; i++ {
		b[i] = tokenChars[seededRand.Intn(len(tokenChars))]
	}
	return string(b)
}
