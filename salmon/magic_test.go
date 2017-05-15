package salmon

import (
	"crypto/rsa"
	"testing"
	"reflect"
	"math/big"
)

var testPublicKeyString = "RSA.mVgY8RN6URBTstndvmUUPb4UZTdwvwmddSKE5z_jvKUEK6yk1u3rrC9yN8k6FilGj9K0eeUPe2hf4Pj-5CmHww.AQAB"
var testDataURL = "data:application/magic-public-key," + testPublicKeyString
var testKeyID = "ATyfAWA5nA6s62uvxAZTwyciKnFDtl9hCpzZwMVi0PQ"

var n, _ = big.NewInt(0).SetString("8031283789075196565022891546563591368344944062154100509645398892293433370859891943306439907454883747534493461257620351548796452092307094036643522661681091", 10)

var testPublicKey = &rsa.PublicKey{
	N: n,
	E: 65537,
}

func TestFormatPublicKey(t *testing.T) {
	s, err := FormatPublicKey(testPublicKey)
	if err != nil {
		t.Fatal("Expected no error when formatting public key, got:", err)
	}

	if s != testPublicKeyString {
		t.Errorf("Invalid formatted public key: expected \n%v\n but got \n%v", testPublicKeyString, s)
	}
}

func TestParsePublicKey(t *testing.T) {
	pub, err := ParsePublicKey(testPublicKeyString)
	if err != nil {
		t.Fatal("Expected no error when parsing public key, got:", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("Expected a *rsa.PublicKey, got: %T", pub)
	}

	if !reflect.DeepEqual(testPublicKey, rsaPub) {
		t.Errorf("Invalid public key: expected \n%#v\n but got \n%#v", testPublicKey, rsaPub)
	}
}

func TestPublicKeyDataURL(t *testing.T) {
	s, err := PublicKeyDataURL(testPublicKey)
	if err != nil {
		t.Fatal("Expected no error when getting public key data URL, got:", err)
	}

	if s != testDataURL {
		t.Errorf("Invalid formatted public key: expected \n%v\n but got \n%v", testDataURL, s)
	}
}

func TestPublicKeyID(t *testing.T) {
	s, err := PublicKeyID(testPublicKey)
	if err != nil {
		t.Fatal("Expected no error when getting public key ID, got:", err)
	}

	if s != testKeyID {
		t.Errorf("Invalid key ID: expected %v but got %v", testKeyID, s)
	}
}
