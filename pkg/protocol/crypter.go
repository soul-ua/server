package protocol

import (
	"encoding/base64"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// Sign data with privateKey and return base64 encoded signature
func Sign(data []byte, privateKey *crypto.Key) (string, error) {
	signingKeyRing, err := crypto.NewKeyRing(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create key ring: %w", err)
	}

	pgpSignature, err := signingKeyRing.SignDetached(crypto.NewPlainMessage(data))
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	pgpSignatureBase64 := base64.StdEncoding.EncodeToString(pgpSignature.GetBinary())

	return pgpSignatureBase64, nil
}

// VerifySignArmor in base64 of data with armored publicKey
func VerifySignArmor(data []byte, pgpSignatureBase64 string, pubKeyArmor string) error {
	pgpSignatureRaw, err := base64.StdEncoding.DecodeString(pgpSignatureBase64)
	if err != nil {
		return fmt.Errorf("failed to decode pgp signature from base64: %w", err)
	}

	publicKeyObj, err := crypto.NewKeyFromArmored(pubKeyArmor)
	if err != nil {
		panic(err)
	}
	signingKeyRing, err := crypto.NewKeyRing(publicKeyObj)
	if err != nil {
		panic(err)
	}

	pgpSignature := crypto.NewPGPSignature(pgpSignatureRaw)

	return signingKeyRing.VerifyDetached(
		crypto.NewPlainMessage(data),
		pgpSignature,
		crypto.GetUnixTime(),
	)
}
