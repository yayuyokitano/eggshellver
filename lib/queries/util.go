package queries

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

func GetArray(query url.Values, key string) []string {
	if query.Get(key) == "" {
		return []string{}
	}
	return strings.Split(query.Get(key), ",")
}

func fetchTransaction() (tx pgx.Tx, err error) {
	tx = services.Tx
	if tx == nil {
		tx, err = services.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	}
	return
}

func commitTransaction(tx pgx.Tx) (err error) {
	if services.IsTesting {
		return
	}
	err = tx.Commit(context.Background())
	return
}

func GenerateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s / 2) //only works for even, thats fine.
	return hex.EncodeToString(b), err
}

func generateRandomBytes(n int) (b []byte, err error) {
	b = make([]byte, n)
	_, err = rand.Read(b)
	return
}
