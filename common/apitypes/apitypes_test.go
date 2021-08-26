package apitypes

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/chainbing/node/common"
	dbUtils "github.com/chainbing/node/db"
	"github.com/iden3/go-iden3-crypto/babyjub"

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

func TestBigIntStrScannerValuer(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type bigInMeddlerStruct struct {
		I *big.Int `meddler:"i,bigint"` // note the bigint that instructs meddler to use BigIntMeddler
	}
	type bigIntStrStruct struct {
		I BigIntStr `meddler:"i"` // note that no meddler is specified, and Scan/Value will be used
	}
	type bigInMeddlerStructNil struct {
		I *big.Int `meddler:"i,bigintnull"` // note the bigint that instructs meddler to use BigIntNullMeddler
	}
	type bigIntStrStructNil struct {
		I *BigIntStr `meddler:"i"` // note that no meddler is specified, and Scan/Value will be used
	}

	// Not nil case
	// Insert into DB using meddler
	const x = int64(12345)
	fromMeddler := bigInMeddlerStruct{
		I: big.NewInt(x),
	}
	err = meddler.Insert(db, "test", &fromMeddler)
	assert.NoError(t, err)
	// Read from DB using BigIntStr
	toBigIntStr := bigIntStrStruct{}
	err = meddler.QueryRow(db, &toBigIntStr, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromMeddler.I.String(), string(toBigIntStr.I))
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using BigIntStr
	fromBigIntStr := bigIntStrStruct{
		I: "54321",
	}
	err = meddler.Insert(db, "test", &fromBigIntStr)
	assert.NoError(t, err)
	// Read from DB using meddler
	toMeddler := bigInMeddlerStruct{}
	err = meddler.QueryRow(db, &toMeddler, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, string(fromBigIntStr.I), toMeddler.I.String())

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using meddler
	fromMeddlerNil := bigInMeddlerStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromMeddlerNil)
	assert.NoError(t, err)
	// Read from DB using BigIntStr
	foo := BigIntStr("foo")
	toBigIntStrNil := bigIntStrStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toBigIntStrNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toBigIntStrNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using BigIntStr
	fromBigIntStrNil := bigIntStrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromBigIntStrNil)
	assert.NoError(t, err)
	// Read from DB using meddler
	toMeddlerNil := bigInMeddlerStructNil{
		I: big.NewInt(x), // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toMeddlerNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toMeddlerNil.I)
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

func TestStrCbBJJ(t *testing.T) {
	type testStrCbBJJ struct {
		I StrCbBJJ
	}
	priv := babyjub.NewRandPrivKey()
	cbBjj := NewCbBJJ(priv.Public().Compress())
	from := []byte(`{"I":"` + cbBjj + `"}`)
	to := &testStrCbBJJ{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, priv.Public().Compress(), (babyjub.PublicKeyComp)(to.I))
}

func TestStrCbIdx(t *testing.T) {
	type testStrCbIdx struct {
		I StrCbIdx
	}
	from := []byte(`{"I":"cb:foo:4"}`)
	to := &testStrCbIdx{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, common.Idx(4), common.Idx(to.I))
}

func TestCbEthAddr(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type ethAddrStruct struct {
		I ethCommon.Address `meddler:"i"`
	}
	type cbEthAddrStruct struct {
		I CbEthAddr `meddler:"i"`
	}
	type ethAddrStructNil struct {
		I *ethCommon.Address `meddler:"i"`
	}
	type cbEthAddrStructNil struct {
		I *CbEthAddr `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using ethCommon.Address Scan/Value
	fromEth := ethAddrStruct{
		I: ethCommon.BigToAddress(big.NewInt(73737373)),
	}
	err = meddler.Insert(db, "test", &fromEth)
	assert.NoError(t, err)
	// Read from DB using CbEthAddr Scan/Value
	toCbEth := cbEthAddrStruct{}
	err = meddler.QueryRow(db, &toCbEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewCbEthAddr(fromEth.I), toCbEth.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using CbEthAddr Scan/Value
	fromCbEth := cbEthAddrStruct{
		I: NewCbEthAddr(ethCommon.BigToAddress(big.NewInt(3786872586))),
	}
	err = meddler.Insert(db, "test", &fromCbEth)
	assert.NoError(t, err)
	// Read from DB using ethCommon.Address Scan/Value
	toEth := ethAddrStruct{}
	err = meddler.QueryRow(db, &toEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromCbEth.I, NewCbEthAddr(toEth.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using ethCommon.Address Scan/Value
	fromEthNil := ethAddrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromEthNil)
	assert.NoError(t, err)
	// Read from DB using CbEthAddr Scan/Value
	foo := CbEthAddr("foo")
	toCbEthNil := cbEthAddrStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toCbEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toCbEthNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using CbEthAddr Scan/Value
	fromCbEthNil := cbEthAddrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromCbEthNil)
	assert.NoError(t, err)
	// Read from DB using ethCommon.Address Scan/Value
	fooAddr := ethCommon.BigToAddress(big.NewInt(1))
	toEthNil := ethAddrStructNil{
		I: &fooAddr, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toEthNil.I)
}

func TestCbBJJ(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type bjjStruct struct {
		I babyjub.PublicKeyComp `meddler:"i"`
	}
	type cbBJJStruct struct {
		I CbBJJ `meddler:"i"`
	}
	type bjjStructNil struct {
		I *babyjub.PublicKeyComp `meddler:"i"`
	}
	type cbBJJStructNil struct {
		I *CbBJJ `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using *babyjub.PublicKeyComp Scan/Value
	priv := babyjub.NewRandPrivKey()
	fromBJJ := bjjStruct{
		I: priv.Public().Compress(),
	}
	err = meddler.Insert(db, "test", &fromBJJ)
	assert.NoError(t, err)
	// Read from DB using CbBJJ Scan/Value
	toCbBJJ := cbBJJStruct{}
	err = meddler.QueryRow(db, &toCbBJJ, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewCbBJJ(fromBJJ.I), toCbBJJ.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using CbBJJ Scan/Value
	fromCbBJJ := cbBJJStruct{
		I: NewCbBJJ(priv.Public().Compress()),
	}
	err = meddler.Insert(db, "test", &fromCbBJJ)
	assert.NoError(t, err)
	// Read from DB using *babyjub.PublicKeyComp Scan/Value
	toBJJ := bjjStruct{}
	err = meddler.QueryRow(db, &toBJJ, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromCbBJJ.I, NewCbBJJ(toBJJ.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using *babyjub.PublicKeyComp Scan/Value
	fromBJJNil := bjjStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromBJJNil)
	assert.NoError(t, err)
	// Read from DB using CbBJJ Scan/Value
	foo := CbBJJ("foo")
	toCbBJJNil := cbBJJStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toCbBJJNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toCbBJJNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using CbBJJ Scan/Value
	fromCbBJJNil := cbBJJStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromCbBJJNil)
	assert.NoError(t, err)
	// Read from DB using *babyjub.PublicKeyComp Scan/Value
	bjjComp := priv.Public().Compress()
	toBJJNil := bjjStructNil{
		I: &bjjComp, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toBJJNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toBJJNil.I)
}

func TestEthSignature(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type ethSignStruct struct {
		I []byte `meddler:"i"`
	}
	type cbEthSignStruct struct {
		I EthSignature `meddler:"i"`
	}
	type cbEthSignStructNil struct {
		I *EthSignature `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using []byte Scan/Value
	s := "someRandomFooForYou"
	fromEth := ethSignStruct{
		I: []byte(s),
	}
	err = meddler.Insert(db, "test", &fromEth)
	assert.NoError(t, err)
	// Read from DB using EthSignature Scan/Value
	toCbEth := cbEthSignStruct{}
	err = meddler.QueryRow(db, &toCbEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewEthSignature(fromEth.I), &toCbEth.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using EthSignature Scan/Value
	fromCbEth := cbEthSignStruct{
		I: *NewEthSignature([]byte(s)),
	}
	err = meddler.Insert(db, "test", &fromCbEth)
	assert.NoError(t, err)
	// Read from DB using []byte Scan/Value
	toEth := ethSignStruct{}
	err = meddler.QueryRow(db, &toEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, &fromCbEth.I, NewEthSignature(toEth.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using []byte Scan/Value
	fromEthNil := ethSignStruct{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromEthNil)
	assert.NoError(t, err)
	// Read from DB using EthSignature Scan/Value
	foo := EthSignature("foo")
	toCbEthNil := cbEthSignStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toCbEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toCbEthNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using EthSignature Scan/Value
	fromCbEthNil := cbEthSignStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromCbEthNil)
	assert.NoError(t, err)
	// Read from DB using []byte Scan/Value
	toEthNil := ethSignStruct{
		I: []byte(s), // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toEthNil.I)
}
