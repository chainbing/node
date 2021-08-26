/*
Package apitypes is used to map the common types used across the node with the format expected by the API.

This is done using different strategies:
- Marshallers: they get triggered when the API marshals the response structs into JSONs
- Scanners/Valuers: they get triggered when a struct is sent/received to/from the SQL database
- Adhoc functions: when the already mentioned strategies are not suitable, functions are added to the structs to facilitate the conversions
*/
package apitypes

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/chainbing/node/common"
	"github.com/chainbing/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// BigIntStr is used to scan/value *big.Int directly into strings from/to sql DBs.
// It assumes that *big.Int are inserted/fetched to/from the DB using the BigIntMeddler meddler
// defined at github.com/chainbing/node/db.  Since *big.Int is
// stored as DECIMAL in SQL, there's no need to implement Scan()/Value()
// because DECIMALS are encoded/decoded as strings by the sql driver, and
// BigIntStr is already a string.
type BigIntStr string

// NewBigIntStr creates a *BigIntStr from a *big.Int.
// If the provided bigInt is nil the returned *BigIntStr will also be nil
func NewBigIntStr(bigInt *big.Int) *BigIntStr {
	if bigInt == nil {
		return nil
	}
	bigIntStr := BigIntStr(bigInt.String())
	return &bigIntStr
}

// StrBigInt is used to unmarshal BigIntStr directly into an alias of big.Int
type StrBigInt big.Int

// UnmarshalText unmarshals a StrBigInt
func (s *StrBigInt) UnmarshalText(text []byte) error {
	bi, ok := (*big.Int)(s).SetString(string(text), 10)
	if !ok {
		return tracerr.Wrap(fmt.Errorf("could not unmarshal %s into a StrBigInt", text))
	}
	*s = StrBigInt(*bi)
	return nil
}

// CollectedFeesAPI is send common.batch.CollectedFee through the API
type CollectedFeesAPI map[common.TokenID]BigIntStr

// NewCollectedFeesAPI creates a new CollectedFeesAPI from a *big.Int map
func NewCollectedFeesAPI(m map[common.TokenID]*big.Int) CollectedFeesAPI {
	c := CollectedFeesAPI(make(map[common.TokenID]BigIntStr))
	for k, v := range m {
		c[k] = *NewBigIntStr(v)
	}
	return c
}

// CbEthAddr is used to scan/value Ethereum Address directly into strings that follow the Ethereum address cb format (^cb:0x[a-fA-F0-9]{40}$) from/to sql DBs.
// It assumes that Ethereum Address are inserted/fetched to/from the DB using the default Scan/Value interface
type CbEthAddr string

// NewCbEthAddr creates a CbEthAddr from an Ethereum addr
func NewCbEthAddr(addr ethCommon.Address) CbEthAddr {
	return CbEthAddr("cb:" + addr.String())
}

// ToEthAddr returns an Ethereum Address created from CbEthAddr
func (a CbEthAddr) ToEthAddr() (ethCommon.Address, error) {
	addrStr := strings.TrimPrefix(string(a), "cb:")
	var addr ethCommon.Address
	return addr, addr.UnmarshalText([]byte(addrStr))
}

// Scan implements Scanner for database/sql
func (a *CbEthAddr) Scan(src interface{}) error {
	ethAddr := &ethCommon.Address{}
	if err := ethAddr.Scan(src); err != nil {
		return tracerr.Wrap(err)
	}
	if ethAddr == nil {
		return nil
	}
	*a = NewCbEthAddr(*ethAddr)
	return nil
}

// Value implements valuer for database/sql
func (a CbEthAddr) Value() (driver.Value, error) {
	ethAddr, err := a.ToEthAddr()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return ethAddr.Value()
}

// StrCbEthAddr is used to unmarshal CbEthAddr directly into an alias of ethCommon.Address
type StrCbEthAddr ethCommon.Address

// UnmarshalText unmarshals a StrCbEthAddr
func (s *StrCbEthAddr) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = StrCbEthAddr(common.EmptyAddr)
		return nil
	}
	withoutCb := strings.TrimPrefix(string(text), "cb:")
	var addr ethCommon.Address
	if err := addr.UnmarshalText([]byte(withoutCb)); err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrCbEthAddr(addr)
	return nil
}

// CbBJJ is used to scan/value *babyjub.PublicKeyComp directly into strings that follow the BJJ public key cb format (^cb:[A-Za-z0-9_-]{44}$) from/to sql DBs.
// It assumes that *babyjub.PublicKeyComp are inserted/fetched to/from the DB using the default Scan/Value interface
type CbBJJ string

// NewCbBJJ creates a CbBJJ from a *babyjub.PublicKeyComp.
// Calling this method with a nil bjj causes panic
func NewCbBJJ(pkComp babyjub.PublicKeyComp) CbBJJ {
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return CbBJJ("cb:" + base64.RawURLEncoding.EncodeToString(bjjSum))
}

func cbStrToBJJ(s string) (babyjub.PublicKeyComp, error) {
	const decodedLen = 33
	const encodedLen = 44
	formatErr := errors.New("invalid BJJ format. Must follow this regex: ^cb:[A-Za-z0-9_-]{44}$")
	encoded := strings.TrimPrefix(s, "cb:")
	if len(encoded) != encodedLen {
		return common.EmptyBJJComp, formatErr
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return common.EmptyBJJComp, formatErr
	}
	if len(decoded) != decodedLen {
		return common.EmptyBJJComp, formatErr
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return common.EmptyBJJComp, tracerr.Wrap(errors.New("checksum verification failed"))
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return bjjComp, nil
}

// ToBJJ returns a babyjub.PublicKeyComp created from CbBJJ
func (b CbBJJ) ToBJJ() (babyjub.PublicKeyComp, error) {
	return cbStrToBJJ(string(b))
}

// Scan implements Scanner for database/sql
func (b *CbBJJ) Scan(src interface{}) error {
	bjj := &babyjub.PublicKeyComp{}
	if err := bjj.Scan(src); err != nil {
		return tracerr.Wrap(err)
	}
	if bjj == nil {
		return nil
	}
	*b = NewCbBJJ(*bjj)
	return nil
}

// Value implements valuer for database/sql
func (b CbBJJ) Value() (driver.Value, error) {
	bjj, err := b.ToBJJ()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return bjj.Value()
}

// StrCbBJJ is used to unmarshal CbBJJ directly into an alias of babyjub.PublicKeyComp
type StrCbBJJ babyjub.PublicKeyComp

// UnmarshalText unmarshalls a StrCbBJJ
func (s *StrCbBJJ) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = StrCbBJJ(common.EmptyBJJComp)
		return nil
	}
	bjj, err := cbStrToBJJ(string(text))
	if err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrCbBJJ(bjj)
	return nil
}

// CbIdx is used to value common.Idx directly into strings that follow the Idx key cb format (cb:tokenSymbol:idx) to sql DBs.
// Note that this can only be used to insert to DB since there is no way to automatically read from the DB since it needs the tokenSymbol
type CbIdx string

// StrCbIdx is used to unmarshal CbIdx directly into an alias of common.Idx
type StrCbIdx common.Idx

// UnmarshalText unmarshals a StrCbIdx
func (s *StrCbIdx) UnmarshalText(text []byte) error {
	withoutCb := strings.TrimPrefix(string(text), "cb:")
	splitted := strings.Split(withoutCb, ":")
	const expectedLen = 2
	if len(splitted) != expectedLen {
		return tracerr.Wrap(fmt.Errorf("can not unmarshal %s into StrCbIdx", text))
	}
	idxInt, err := strconv.Atoi(splitted[1])
	if err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrCbIdx(common.Idx(idxInt))
	return nil
}

// EthSignature is used to scan/value []byte representing an Ethereum signature directly into strings from/to sql DBs.
type EthSignature string

// NewEthSignature creates a *EthSignature from []byte
// If the provided signature is nil the returned *EthSignature will also be nil
func NewEthSignature(signature []byte) *EthSignature {
	if signature == nil {
		return nil
	}
	ethSignature := EthSignature("0x" + hex.EncodeToString(signature))
	return &ethSignature
}

// Scan implements Scanner for database/sql
func (e *EthSignature) Scan(src interface{}) error {
	if srcStr, ok := src.(string); ok {
		// src is a string
		*e = *(NewEthSignature([]byte(srcStr)))
		return nil
	} else if srcBytes, ok := src.([]byte); ok {
		// src is []byte
		*e = *(NewEthSignature(srcBytes))
		return nil
	} else {
		// unexpected src
		return tracerr.Wrap(fmt.Errorf("can't scan %T into apitypes.EthSignature", src))
	}
}

// Value implements valuer for database/sql
func (e EthSignature) Value() (driver.Value, error) {
	without0x := strings.TrimPrefix(string(e), "0x")
	return hex.DecodeString(without0x)
}

// UnmarshalText unmarshals a StrEthSignature
func (e *EthSignature) UnmarshalText(text []byte) error {
	without0x := strings.TrimPrefix(string(text), "0x")
	signature, err := hex.DecodeString(without0x)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*e = EthSignature([]byte(signature))
	return nil
}
