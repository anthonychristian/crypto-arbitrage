package indodax

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

func sign(dat interface{}, secret string) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha512.New, []byte(secret))

	// Write Data to it
	_, _ = h.Write([]byte(dat.(string)))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}
