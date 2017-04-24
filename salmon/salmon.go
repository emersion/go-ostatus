// Package salmon implements the Salmon protocol, as defined in
// http://www.salmon-protocol.org/salmon-protocol-summary.
package salmon

import (
	"encoding/xml"
)

// TODO: JSON schema

type MagicEnv struct {
	XMLName  xml.Name    `xml:"http://salmon-protocol.org/ns/magic-env env" json:"-"`
	Data     *MagicData  `xml:"data" json:"data"`
	Encoding string      `xml:"encoding" json:"encoding"`
	Alg      string      `xml:"alg" json:"alg"`
	Sig      []*MagicSig `xml:"sig" json:"sigs"`
}

type MagicData struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type MagicSig struct {
	KeyID string `xml:"key_id,attr" json:"key_id"`
	Value string `xml:",chardata" json:"value"`
}
