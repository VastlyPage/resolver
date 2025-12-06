package hlutil

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/net/idna"
	"github.com/adraffy/go-ens-normalize/ensip15"
)

// https://eips.ethereum.org/EIPS/eip-137#namehash-algorithm
func nameHash(name string) [32]byte {
	var node [32]byte // Start with 32 bytes of 0
	if name == "" {
		return node
	}

	labels := strings.Split(name, ".")
	for i := len(labels) - 1; i >= 0; i-- {
		labelHash := crypto.Keccak256([]byte(labels[i]))
		copy(node[:], crypto.Keccak256Hash(node[:], labelHash).Bytes())
	}
	return node
}

func NameHash(name string) string {
	return fmt.Sprintf("0x%x", nameHash(name))
}

func EnsureHTTPPrefix(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func GetStringOrEmpty(kv map[string]interface{}, key string) string {
	if val, ok := kv[key].(string); ok {
		return val
	}
	return ""
}

func Normalize(real string) string {
	unicode, _ := idna.ToUnicode(real)
	normalized := ensip15.Normalize(unicode + ".hl")
	return normalized
}
