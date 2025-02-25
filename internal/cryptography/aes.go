package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

func Encrypt(key []byte, plaintext []byte) ([]byte, error) {
	encryptor, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(encryptor)
	if err != nil {
		return nil, err
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	encryptor, err := aes.NewCipher(key)
	if err != nil {
		return make([]byte, 0), err
	}

	gcm, err := cipher.NewGCM(encryptor)
	if err != nil {
		return make([]byte, 0), err
	}

	// Since we know the ciphertext is actually nonce+ciphertext
	// And len(nonce) == NonceSize(). We can separate the two.
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return make([]byte, 0), err
	}

	return plaintext, err
}
