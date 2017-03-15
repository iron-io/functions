package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"github.com/iron-io/functions/api/models"
	"io"
)

type DockerAuth struct {
	Datastore models.Datastore
	Key       []byte
	Nonce     []byte
}

func NewDockerMock(ds models.Datastore) DockerAuth {
	return DockerAuth{
		Datastore: ds,
		Key:       []byte("A159B69FAF460F55C0966B6383CE0917"),
	}
}

func (da *DockerAuth) SaveDockerCredentials(ctx context.Context, dockerLogin models.DockerCreds) error {

	val, err := json.Marshal(dockerLogin)
	if err != nil {
		return err
	}
	encryptedData, nonce := da.encrypt(val)

	err = da.Datastore.Put(ctx, []byte("dockerLogin-nonce"), nonce)
	if err != nil {
		return err
	}

	return da.Datastore.Put(ctx, []byte("dockerLogin"), encryptedData)
}

func (da *DockerAuth) GetAuthConfiguration(ctx context.Context) (*docker.AuthConfiguration, error) {

	data, err := da.Datastore.Get(ctx, []byte("dockerLogin"))
	if err != nil {
		return nil, err
	}
	nonce, err := da.Datastore.Get(ctx, []byte("dockerLogin-nonce"))
	if err != nil {
		return nil, err
	}
	if data != nil {
		data = da.decrypt(data, nonce)
		creds := &models.DockerCreds{}
		err = json.Unmarshal(data, creds)
		if err != nil {
			return nil, err
		}
		return creds.ToDockerAuthentication()
	} else {
		return &docker.AuthConfiguration{}, nil
	}
}

func (da *DockerAuth) encrypt(data []byte) ([]byte, []byte) {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	block, err := aes.NewCipher(da.Key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	return ciphertext, nonce
}

func (da *DockerAuth) decrypt(ciphertext, nonce []byte) []byte {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.

	block, err := aes.NewCipher(da.Key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return plaintext
}
