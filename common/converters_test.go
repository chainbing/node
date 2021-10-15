package common

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	dbUtils "github.com/chainbing/node/db"

	// nolint sqlite driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/meddler"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Register meddler
	meddler.Default = meddler.SQLite
	meddler.Register("bigint", dbUtils.BigIntMeddler{})
	meddler.Register("bigintnull", dbUtils.BigIntNullMeddler{})
	// Create temporary sqlite DB
	dir, err := ioutil.TempDir("", "db")
	if err != nil {
		panic(err)
	}
	db, err = sql.Open("sqlite3", dir+"sqlite.db")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir) //nolint
	schema := `CREATE TABLE test (i BLOB);`
	if _, err := db.Exec(schema); err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	os.Exit(result)
}

func TestStrBigInt(t *testing.T) {
	type testStrBigInt struct {
		I StrBigInt
	}
	from := []byte(`{"I":"4"}`)
	to := &testStrBigInt{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, big.NewInt(4), (*big.Int)(&to.I))
}

func TestStrCbEthAddr(t *testing.T) {
	type testStrCbEthAddr struct {
		I StrCbEthAddr
	}
	withoutCb := "0xaa942cfcd25ad4d90a62358b0dd84f33b398262a"
	from := []byte(`{"I":"cb:` + withoutCb + `"}`)
	var addr ethCommon.Address
	if err := addr.UnmarshalText([]byte(withoutCb)); err != nil {
		panic(err)
	}
	to := &testStrCbEthAddr{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, addr, ethCommon.Address(to.I))
}

func TestStrCbIdx(t *testing.T) {
	type testStrCbIdx struct {
		I StrCbIdx
	}
	from := []byte(`{"I":"cb:foo:4"}`)
	to := &testStrCbIdx{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, Idx(4), Idx(to.I.Idx))
	assert.Equal(t, "foo", to.I.TokenSymbol)
}
