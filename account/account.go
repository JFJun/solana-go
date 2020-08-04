package account

import (
	"crypto/ed25519"
	"crypto/rand"
	"github.com/btcsuite/btcutil/base58"
)

type Account struct {
	SecretKey []byte
	PublicKey []byte
}

func (acc *Account) ToBase58() string {
	if acc.PublicKey == nil {
		return ""
	}
	return base58.Encode(acc.PublicKey)
}

func NewAccount() (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	acc := new(Account)
	acc.SecretKey = priv.Seed()
	acc.PublicKey = pub
	return acc, nil
}
func NewAccountBySecret(seed []byte) *Account {
	priv := ed25519.NewKeyFromSeed(seed)
	acc := new(Account)
	acc.SecretKey = priv.Seed()
	acc.PublicKey = priv[32:]
	return acc
}
