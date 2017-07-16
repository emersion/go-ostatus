package salmon

import (
	"crypto/rsa"
	"encoding/xml"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
)

const testReply = `<?xml version='1.0' encoding='UTF-8'?>
<entry xmlns='http://www.w3.org/2005/Atom'>
  <id>tag:example.com,2009:cmt-0.44775718</id>
  <author><name>test@example.com</name><uri>bob@example.com</uri></author>
  <thr:in-reply-to xmlns:thr='http://purl.org/syndication/thread/1.0'
      ref='tag:blogger.com,1999:blog-893591374313312737.post-3861663258538857954'>tag:blogger.com,1999:blog-893591374313312737.post-3861663258538857954
  </thr:in-reply-to>
  <content>Salmon swim upstream!</content>
  <title>Salmon swim upstream!</title>
  <updated>2009-12-18T20:04:03Z</updated>
</entry>
`

const testMagicEnvXML = `<?xml version='1.0' encoding='UTF-8'?>
<me:env xmlns:me='http://salmon-protocol.org/ns/magic-env'>
  <me:data type='application/atom+xml'>
    PD94bWwgdmVyc2lvbj0nMS4wJyBlbmNvZGluZz0nVVRGLTgnPz4KPGVudHJ5IHhtbG5zPSdod
    HRwOi8vd3d3LnczLm9yZy8yMDA1L0F0b20nPgogIDxpZD50YWc6ZXhhbXBsZS5jb20sMjAwOT
    pjbXQtMC40NDc3NTcxODwvaWQ-ICAKICA8YXV0aG9yPjxuYW1lPnRlc3RAZXhhbXBsZS5jb20
    8L25hbWU-PHVyaT5ib2JAZXhhbXBsZS5jb208L3VyaT48L2F1dGhvcj4KICA8dGhyOmluLXJl
    cGx5LXRvIHhtbG5zOnRocj0naHR0cDovL3B1cmwub3JnL3N5bmRpY2F0aW9uL3RocmVhZC8xL
    jAnCiAgICAgIHJlZj0ndGFnOmJsb2dnZXIuY29tLDE5OTk6YmxvZy04OTM1OTEzNzQzMTMzMT
    I3MzcucG9zdC0zODYxNjYzMjU4NTM4ODU3OTU0Jz50YWc6YmxvZ2dlci5jb20sMTk5OTpibG9
    nLTg5MzU5MTM3NDMxMzMxMjczNy5wb3N0LTM4NjE2NjMyNTg1Mzg4NTc5NTQKICA8L3Rocjpp
    bi1yZXBseS10bz4KICA8Y29udGVudD5TYWxtb24gc3dpbSB1cHN0cmVhbSE8L2NvbnRlbnQ-C
    iAgPHRpdGxlPlNhbG1vbiBzd2ltIHVwc3RyZWFtITwvdGl0bGU-CiAgPHVwZGF0ZWQ-MjAwOS
    0xMi0xOFQyMDowNDowM1o8L3VwZGF0ZWQ-CjwvZW50cnk-CiAgICA=
  </me:data>
  <me:encoding>base64url</me:encoding>
  <me:alg>RSA-SHA256</me:alg>
  <me:sig>
    cAIu8VKIhs3WedN91L3ynLT3GbZFhbVidDn-skGetENVH-3EguaYIjlPTq7Ieraq4SD
    BknM9STM9DR90kveUrw==
  </me:sig>
</me:env>
`

const testMagicEnvJSON = `{
  "data": "PD94bWwgdmVyc2lvbj0nMS4wJyBlbmNvZGluZz0nVVRGLTgnPz4KPGVudHJ5IHhtbG5zPSdodHRwOi8vd3d3LnczLm9yZy8yMDA1L0F0b20nPgogIDxpZD50YWc6ZXhhbXBsZS5jb20sMjAwOTpjbXQtMC40NDc3NTcxODwvaWQ-ICAKICA8YXV0aG9yPjxuYW1lPnRlc3RAZXhhbXBsZS5jb208L25hbWU-PHVyaT5ib2JAZXhhbXBsZS5jb208L3VyaT48L2F1dGhvcj4KICA8dGhyOmluLXJlcGx5LXRvIHhtbG5zOnRocj0naHR0cDovL3B1cmwub3JnL3N5bmRpY2F0aW9uL3RocmVhZC8xLjAnCiAgICAgIHJlZj0ndGFnOmJsb2dnZXIuY29tLDE5OTk6YmxvZy04OTM1OTEzNzQzMTMzMTI3MzcucG9zdC0zODYxNjYzMjU4NTM4ODU3OTU0Jz50YWc6YmxvZ2dlci5jb20sMTk5OTpibG9nLTg5MzU5MTM3NDMxMzMxMjczNy5wb3N0LTM4NjE2NjMyNTg1Mzg4NTc5NTQKICA8L3Rocjppbi1yZXBseS10bz4KICA8Y29udGVudD5TYWxtb24gc3dpbSB1cHN0cmVhbSE8L2NvbnRlbnQ-CiAgPHRpdGxlPlNhbG1vbiBzd2ltIHVwc3RyZWFtITwvdGl0bGU-CiAgPHVwZGF0ZWQ-MjAwOS0xMi0xOFQyMDowNDowM1o8L3VwZGF0ZWQ-CjwvZW50cnk-CiAgICA=",
  "data_type": "application/atom+xml",
  "encoding": "base64url",
  "alg": "RSA-SHA256",
  "sigs": [
    {
    "value": "EvGSD2vi8qYcveHnb-rrlok07qnCXjn8YSeCDDXlbhILSabgvNsPpbe76up8w63i2fWHvLKJzeGLKfyHg8ZomQ",
    "key_id": "4k8ikoyC2Xh+8BiIeQ+ob7Hcd2J7/Vj3uM61dy9iRMI="
    }
  ]
}`

func TestMagicEnv(t *testing.T) {
	// Generate an insecure test key - we don't care
	priv, err := rsa.GenerateKey(rand.New(rand.NewSource(0)), 512)
	if err != nil {
		t.Fatal("Cannot generate private key:", err)
	}

	env, err := CreateMagicEnv("application/atom+xml", []byte(testReply), priv)
	if err != nil {
		t.Fatalf("CreateMagicEnv() = %v", err)
	}

	if err := env.Verify(&priv.PublicKey); err != nil {
		t.Errorf("Verify(correct key) = %v", err)
	}
	if err := env.Verify(testPublicKey); err == nil {
		t.Errorf("Verify(incorrect key) = %v", err)
	}
}

func TestMagicEnv_xml(t *testing.T) {
	r := strings.NewReader(testMagicEnvXML)

	env := new(MagicEnv)
	if err := xml.NewDecoder(r).Decode(env); err != nil {
		t.Fatal("Expected no error when parsing magic envelope, got:", err)
	}

	b, err := env.UnverifiedData()
	if err != nil {
		t.Fatalf("UnverifiedData() = %v", err)
	}

	s := strings.Replace(string(b), "  \n", "\n", -1)
	s = strings.Trim(s, " ")
	if s != testReply {
		t.Errorf("UnverifiedData() = \n%v\n, want \n%v", s, testReply)
	}

	// TODO
	//if err := env.Verify(pub); err != nil {
	//	t.Fatal("Verify() = ", err)
	//}
}

func TestMagicEnv_json(t *testing.T) {
	r := strings.NewReader(testMagicEnvJSON)

	env := new(MagicEnv)
	if err := json.NewDecoder(r).Decode(env); err != nil {
		t.Fatal("Expected no error when parsing magic envelope, got:", err)
	}

	b, err := env.UnverifiedData()
	if err != nil {
		t.Fatalf("UnverifiedData() = %v", err)
	}

	s := strings.Replace(string(b), "  \n", "\n", -1)
	s = strings.Trim(s, " ")
	if s != testReply {
		t.Errorf("UnverifiedData() = \n%v\n, want \n%v", s, testReply)
	}

	// TODO
	//if err := env.Verify(pub); err != nil {
	//	t.Fatal("Verify() = ", err)
	//}
}
