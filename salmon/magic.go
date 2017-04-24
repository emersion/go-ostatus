package salmon

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"strings"
)

var enc = base64.URLEncoding

var (
	errUnknownPublicKeyType = errors.New("salmon: unknown public key type")
	errMalformedPublicKey = errors.New("salmon: malformed public key")
)

func FormatPublicKey(pk crypto.PublicKey) (string, error) {
	switch pk := pk.(type) {
	case *rsa.PublicKey:
		n := enc.EncodeToString(pk.N.Bytes())
		e := enc.EncodeToString(big.NewInt(int64(pk.E)).Bytes())
		return "RSA."+n+"."+e, nil
	default:
		return "", errUnknownPublicKeyType
	}
}

func ParsePublicKey(s string) (crypto.PublicKey, error) {
	parts := strings.Split(s, ".")
	switch strings.ToUpper(parts[0]) {
	case "RSA":
		if len(parts) != 3 {
			return nil, errMalformedPublicKey
		}

		n, err := enc.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
		e, err := enc.DecodeString(parts[2])
		if err != nil {
			return nil, err
		}

		return &rsa.PublicKey{
			N: big.NewInt(0).SetBytes(n),
			E: int(big.NewInt(0).SetBytes(e).Int64()),
		}, nil
	default:
		return nil, errUnknownPublicKeyType
	}
}

func PublicKeyDataURL(pk crypto.PublicKey) (string, error) {
	s, err := FormatPublicKey(pk)
	if err != nil {
		return "", err
	}
	return "data:application/magic-public-key,"+s, nil
}
