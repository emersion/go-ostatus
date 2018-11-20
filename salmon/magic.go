package salmon

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"strings"
	"unicode"
)

// MagicPublicKeyRel is the magic-public-key relation.
const RelMagicPublicKey = "magic-public-key"

var (
	errUnknownKeyType       = errors.New("salmon: unknown key type")
	errMalformedPublicKey   = errors.New("salmon: malformed public key")
	errUnknownAlg           = errors.New("salmon: unknown signature algorithm")
	errInvalidPublicKeyType = errors.New("salmon: invalid public key type")
)

func decodeString(s string) ([]byte, error) {
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	// The spec says to use URL encoding without padding, but some implementations
	// add padding (e.g. Mastodon).
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}

func encodeToString(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func signatureEncodeToString(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}

// FormatPublicKey formats a public key into the application/magic-key format.
func FormatPublicKey(pk crypto.PublicKey) (string, error) {
	switch pk := pk.(type) {
	case *rsa.PublicKey:
		n := encodeToString(pk.N.Bytes())
		e := encodeToString(big.NewInt(int64(pk.E)).Bytes())
		return "RSA." + n + "." + e, nil
	default:
		return "", errUnknownKeyType
	}
}

// ParsePublicKey parses a public key from the application/magic-key format.
func ParsePublicKey(s string) (crypto.PublicKey, error) {
	parts := strings.Split(s, ".")
	switch strings.ToUpper(parts[0]) {
	case "RSA":
		if len(parts) != 3 {
			return nil, errMalformedPublicKey
		}

		n, err := decodeString(parts[1])
		if err != nil {
			return nil, err
		}
		e, err := decodeString(parts[2])
		if err != nil {
			return nil, err
		}

		return &rsa.PublicKey{
			N: big.NewInt(0).SetBytes(n),
			E: int(big.NewInt(0).SetBytes(e).Int64()),
		}, nil
	default:
		return nil, errUnknownKeyType
	}
}

const dataURLPrefix = "data:application/magic-public-key,"

// FormatPublicKeyDataURL returns the data URL for a public key.
func FormatPublicKeyDataURL(pk crypto.PublicKey) (string, error) {
	s, err := FormatPublicKey(pk)
	if err != nil {
		return "", err
	}
	return dataURLPrefix + s, nil
}

// ParsePublicKeyDataURL parses a public key data URL.
func ParsePublicKeyDataURL(u string) (crypto.PublicKey, error) {
	// TODO: full data URL support
	if !strings.HasPrefix(u, dataURLPrefix) {
		return nil, errors.New("salmon: not a public key data URL")
	}
	return ParsePublicKey(strings.TrimPrefix(u, dataURLPrefix))
}

// PublicKeyID returns the key identifier for a public key.
func PublicKeyID(pk crypto.PublicKey) (string, error) {
	s, err := FormatPublicKey(pk)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	io.WriteString(h, s)
	id := encodeToString(h.Sum(nil))
	return id, nil
}

func computeHash(env *MagicEnv) ([]byte, error) {
	mediaType := signatureEncodeToString([]byte(env.Data.Type))
	encoding := signatureEncodeToString([]byte(env.Encoding))
	alg := signatureEncodeToString([]byte(env.Alg))

	h := sha256.New()
	_, err := io.WriteString(h, env.Data.Value+"."+mediaType+"."+encoding+"."+alg)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func sign(env *MagicEnv, priv crypto.PrivateKey) error {
	switch priv := priv.(type) {
	case *rsa.PrivateKey:
		if env.Alg != "" && env.Alg != "RSA-SHA256" {
			return errors.New("salmon: cannot sign an envelope with two different algorithms")
		}
		env.Alg = "RSA-SHA256"

		hashed, err := computeHash(env)
		if err != nil {
			return err
		}

		sigb, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
		if err != nil {
			return err
		}

		keyid, err := PublicKeyID(&priv.PublicKey)
		if err != nil {
			return err
		}

		env.Sig = append(env.Sig, &MagicSig{
			KeyID: keyid,
			Value: encodeToString(sigb),
		})
		return nil
	}
	return errUnknownKeyType
}

func verify(env *MagicEnv, pk crypto.PublicKey, sig string) error {
	sigb, err := decodeString(sig)
	if err != nil {
		return err
	}

	hashed, err := computeHash(env)
	if err != nil {
		return err
	}

	switch strings.ToUpper(env.Alg) {
	case "RSA-SHA256":
		pk, ok := pk.(*rsa.PublicKey)
		if !ok {
			return errInvalidPublicKeyType
		}
		return rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed, sigb)
	}
	return errUnknownAlg
}
