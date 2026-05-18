// Package wgkeys генерирует ключи WireGuard без внешней команды wg.
// Это удобно на Windows при локальной разработке.
package wgkeys

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// Pair — пара ключей WireGuard в base64 (как в wg genkey / wg pubkey).
type Pair struct {
	Private string
	Public  string
}

// Generate создаёт новую пару ключей по правилам WireGuard/Curve25519.
func Generate() (Pair, error) {
	var privateKey [32]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return Pair{}, fmt.Errorf("random: %w", err)
	}

	// Clamping — стандарт WireGuard; без этого ключ может быть невалидным.
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return Pair{
		Private: encodeKey(privateKey[:]),
		Public:  encodeKey(publicKey[:]),
	}, nil
}

func encodeKey(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}
