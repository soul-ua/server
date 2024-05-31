package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"log"
)

func GeneratePair(name, email string) (string, string, error) {
	ecKey, err := crypto.GenerateKey(name, email, "x25519", 0)
	log.Println(ecKey, err)

	armor, err := ecKey.Armor()
	if err != nil {
		panic(err)
	}

	armorPub, err := ecKey.GetArmoredPublicKey()
	if err != nil {
		panic(err)
	}

	return armor, armorPub, nil
}

// Sign data with privateKey and return base64 encoded signature
func Sign(data []byte, privateKey *crypto.Key) (string, error) {
	signingKeyRing, err := crypto.NewKeyRing(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed To create key ring: %w", err)
	}

	pgpSignature, err := signingKeyRing.SignDetached(crypto.NewPlainMessage(data))
	if err != nil {
		return "", fmt.Errorf("failed To sign data: %w", err)
	}

	pgpSignatureBase64 := base64.StdEncoding.EncodeToString(pgpSignature.GetBinary())

	return pgpSignatureBase64, nil
}

// VerifySignArmor in base64 of data with armored publicKey
func VerifySignArmor(data []byte, pgpSignatureBase64 string, pubKeyArmor string) error {
	pgpSignatureRaw, err := base64.StdEncoding.DecodeString(pgpSignatureBase64)
	if err != nil {
		return fmt.Errorf("failed To decode pgp signature From base64: %w", err)
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

func EncryptStructSign(v interface{}, publicKeyArmor string, privateKeyArmor string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed To marshal data: %w", err)
	}

	return EncryptSign(data, publicKeyArmor, privateKeyArmor)
}

func EncryptSign(data []byte, publicKeyArmor string, privateKeyArmor string) ([]byte, error) {
	publicKey, err := crypto.NewKeyFromArmored(publicKeyArmor)
	if err != nil {
		return nil, fmt.Errorf("failed To decode public key: %w", err)
	}

	privateKey, err := crypto.NewKeyFromArmored(privateKeyArmor)
	if err != nil {
		return nil, fmt.Errorf("failed To decode private key: %w", err)
	}

	encryptionKeyRing, err := crypto.NewKeyRing(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed To create encryption key ring: %w", err)
	}

	signingKeyRing, err := crypto.NewKeyRing(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed To create signing key ring: %w", err)
	}

	encrypted, err := encryptionKeyRing.Encrypt(crypto.NewPlainMessage(data), signingKeyRing)
	if err != nil {
		return nil, fmt.Errorf("failed To encrypt data: %w", err)
	}

	return encrypted.GetBinary(), nil
}

func DecryptStructVerify[T any](data []byte, publicKeyArmor string, privateKeyArmor string) (T, error) {
	var v T

	decrypted, err := DecryptVerify(data, publicKeyArmor, privateKeyArmor)
	if err != nil {
		return v, fmt.Errorf("failed To decrypt data: %w", err)
	}

	err = json.Unmarshal(decrypted, &v)
	if err != nil {
		return v, fmt.Errorf("failed To unmarshal data: %w", err)
	}

	return v, nil
}

func DecryptVerify(data []byte, publicKeyArmor string, privateKeyArmor string) ([]byte, error) {
	publicKey, err := crypto.NewKeyFromArmored(publicKeyArmor)
	if err != nil {
		return nil, fmt.Errorf("failed To decode public key: %w", err)
	}

	privateKey, err := crypto.NewKeyFromArmored(privateKeyArmor)
	if err != nil {
		return nil, fmt.Errorf("failed To decode private key: %w", err)
	}

	encryptionKeyRing, err := crypto.NewKeyRing(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed To create encryption key ring: %w", err)
	}

	signingKeyRing, err := crypto.NewKeyRing(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed To create signing key ring: %w", err)
	}

	decrypted, err := encryptionKeyRing.Decrypt(crypto.NewPGPMessage(data), signingKeyRing, crypto.GetUnixTime())
	if err != nil {
		return nil, fmt.Errorf("failed To decrypt data: %w", err)
	}

	return decrypted.GetBinary(), nil
}
