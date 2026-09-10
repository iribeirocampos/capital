package capital

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// encryptPassword replicates the Capital.com login password encryption
// scheme: the password and the encryption key's timestamp are concatenated
// as "password|timestamp", RSA (PKCS#1 v1.5) encrypted with the
// server-provided public key, and base64-encoded.
func encryptPassword(password, base64Key string, timestamp int64) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("decode encryption key: %w", err)
	}

	pub, err := parseRSAPublicKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("parse encryption key: %w", err)
	}

	plaintext := fmt.Sprintf("%s|%d", password, timestamp)
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("rsa encrypt: %w", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// parseRSAPublicKey accepts either a raw X.509/PKIX DER-encoded key (as
// returned by the Capital.com API) or a PEM-wrapped one.
func parseRSAPublicKey(der []byte) (*rsa.PublicKey, error) {
	if block, _ := pem.Decode(der); block != nil {
		der = block.Bytes
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unexpected public key type %T", pub)
	}
	return rsaPub, nil
}
