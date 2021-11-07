package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"git.denetwork.xyz/DeNet/dfile-secondary-node/logger"
	"git.denetwork.xyz/DeNet/dfile-secondary-node/shared"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pbnjay/memory"
)

var (
	EncryptedPK []byte
	SecretKey   []byte
)

// ====================================================================================

//EncryptAES encrypts data using a provided key.
func EncryptAES(key, data []byte) ([]byte, error) {
	const location = "encryption.encryptAES->"
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// ====================================================================================

//DecryptAES decrypts data using a provided key.
func DecryptAES(key, data []byte) ([]byte, error) {
	const location = "encryption.decryptAES->"
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}
	nonce, encrData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	decrData, err := gcm.Open(nil, nonce, encrData, nil)
	if err != nil {
		return nil, logger.CreateDetails(location, err)
	}

	return decrData, nil
}

// ====================================================================================

//Return N and P scrypt params
func GetScryptParams() (int, int) {
	if shared.TestMode {
		return keystore.LightScryptN, keystore.LightScryptP
	}

	if memory.TotalMemory()/1024/1024 < 1000 {
		return keystore.LightScryptN * 16, keystore.StandardScryptP
	}

	return keystore.StandardScryptN, keystore.StandardScryptP
}
