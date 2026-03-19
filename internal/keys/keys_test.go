package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestGenerateP256KeyPair(t *testing.T) {
	privPEM, pubPEM, err := GenerateP256KeyPair()
	if err != nil {
		t.Fatalf("GenerateP256KeyPair() error: %v", err)
	}

	// Private key should be valid PEM.
	privBlock, _ := pem.Decode(privPEM)
	if privBlock == nil {
		t.Fatal("private key: invalid PEM encoding")
	}
	if privBlock.Type != "PRIVATE KEY" {
		t.Fatalf("private key PEM type = %q, want PRIVATE KEY", privBlock.Type)
	}

	// Should parse as PKCS8 ECDSA key.
	key, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("parse PKCS8 private key: %v", err)
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("private key type = %T, want *ecdsa.PrivateKey", key)
	}
	if ecKey.Curve.Params().Name != "P-256" {
		t.Fatalf("curve = %s, want P-256", ecKey.Curve.Params().Name)
	}

	// Public key should be valid PEM.
	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		t.Fatal("public key: invalid PEM encoding")
	}
	if pubBlock.Type != "PUBLIC KEY" {
		t.Fatalf("public key PEM type = %q, want PUBLIC KEY", pubBlock.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("parse public key: %v", err)
	}
	ecPub, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("public key type = %T, want *ecdsa.PublicKey", pubKey)
	}

	// Public key should match the private key.
	if !ecKey.PublicKey.Equal(ecPub) {
		t.Fatal("public key does not match private key")
	}
}

func TestPublicKeyFromPrivateKeyPEM_PKCS8(t *testing.T) {
	privPEM, expectedPubPEM, err := GenerateP256KeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}

	gotPubPEM, err := PublicKeyFromPrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("PublicKeyFromPrivateKeyPEM() error: %v", err)
	}

	if string(gotPubPEM) != string(expectedPubPEM) {
		t.Fatalf("public key mismatch:\ngot:\n%s\nwant:\n%s", gotPubPEM, expectedPubPEM)
	}
}

func TestPublicKeyFromPrivateKeyPEM_InvalidPEM(t *testing.T) {
	_, err := PublicKeyFromPrivateKeyPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestPublicKeyFromPrivateKeyPEM_WrongKeyType(t *testing.T) {
	// RSA key disguised as PEM — should fail since it's not ECDSA.
	fakeBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte("not a real key"),
	})
	_, err := PublicKeyFromPrivateKeyPEM(fakeBlock)
	if err == nil {
		t.Fatal("expected error for non-ECDSA key, got nil")
	}
}

func TestGenerateP256KeyPair_Uniqueness(t *testing.T) {
	priv1, _, _ := GenerateP256KeyPair()
	priv2, _, _ := GenerateP256KeyPair()

	if string(priv1) == string(priv2) {
		t.Fatal("two generated key pairs should not be identical")
	}
}
