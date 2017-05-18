// Package salmon implements the Salmon protocol, as defined in
// http://www.salmon-protocol.org/salmon-protocol-summary.
package salmon

import (
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"errors"
)

// Rel is the salmon relation.
const Rel = "salmon"

// A MagicEnv is a magic envelope and contains a message bundled along with
// signature(s) for that message.
type MagicEnv struct {
	XMLName  xml.Name    `xml:"http://salmon-protocol.org/ns/magic-env env"`
	Data     *MagicData  `xml:"data"`
	Encoding string      `xml:"encoding"`
	Alg      string      `xml:"alg"`
	Sig      []*MagicSig `xml:"sig"`
}

// CreateMagicEnv creates a new magic envelope.
func CreateMagicEnv(mediaType string, data []byte, priv crypto.PrivateKey) (*MagicEnv, error) {
	env := &MagicEnv{
		Data: &MagicData{
			Type:  mediaType,
			Value: encodeToString(data),
		},
		Encoding: "base64url",
	}

	if err := sign(env, priv); err != nil {
		return nil, err
	}

	return env, nil
}

type magicEnvJSON struct {
	*MagicData
	Encoding string      `json:"encoding"`
	Alg      string      `json:"alg"`
	Sig      []*MagicSig `json:"sigs"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (env *MagicEnv) UnmarshalJSON(b []byte) error {
	var envJSON magicEnvJSON
	if err := json.Unmarshal(b, &envJSON); err != nil {
		return err
	}

	env.Data = envJSON.MagicData
	env.Encoding = envJSON.Encoding
	env.Alg = envJSON.Alg
	env.Sig = envJSON.Sig
	return nil
}

// MarshalJSON implements json.Marshaler.
func (env *MagicEnv) MarshalJSON() ([]byte, error) {
	return json.Marshal(&magicEnvJSON{
		MagicData: env.Data,
		Encoding:  env.Encoding,
		Alg:       env.Alg,
		Sig:       env.Sig,
	})
}

// UnverifiedData returns this envelope's message, without checking the
// signature.
func (env *MagicEnv) UnverifiedData() ([]byte, error) {
	switch env.Encoding {
	case "base64url":
		return decodeString(env.Data.Value)
	default:
		return nil, errors.New("salmon: unknown envelope encoding")
	}
}

// Verify checks that the envelope is signed with pub.
func (env *MagicEnv) Verify(pub crypto.PublicKey) error {
	if len(env.Sig) == 0 {
		return errors.New("salmon: no signature in envelope")
	}

	var err error
	for _, sig := range env.Sig {
		if err = verify(env, pub, sig.Value); err == nil {
			return nil
		} else if err != rsa.ErrVerification && err != errInvalidPublicKeyType {
			break
		}
	}

	return err
}

// A MagicaData contains a type and a value.
type MagicData struct {
	Type  string `xml:"type,attr" json:"data_type"`
	Value string `xml:",chardata" json:"data"`
}

// A MagicSig is a magic signature.
type MagicSig struct {
	KeyID string `xml:"key_id,attr" json:"key_id"`
	Value string `xml:",chardata" json:"value"`
}
