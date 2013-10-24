package signature


import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSimpleEncryption(t *testing.T) {
	key := "aaaaaaaaaaaaaaaa"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	decrypted, err := Decrypt(key, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, decrypted, message)
	assert.NotEqual(t, encrypted, message)
}

func TestEncryptionWithAShortKey(t *testing.T) {
	key := "short key"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	decrypted, err := Decrypt(key, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, decrypted, message)
	assert.NotEqual(t, encrypted, message)
}

func TestDecryptionWithWrongKey(t *testing.T) {
	key := "short key"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	_, err = Decrypt("wrong key", encrypted)
	assert.Error(t, err)
}
