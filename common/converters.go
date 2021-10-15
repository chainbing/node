package common

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/chainbing/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// StringToTxType converts string to transaction type
func StringToTxType(txType string) (*TxType, error) {
	if txType == "" {
		return nil, nil
	}
	txTypeCasted := TxType(txType)
	switch txTypeCasted {
	case TxTypeExit, TxTypeTransfer, TxTypeDeposit, TxTypeCreateAccountDeposit,
		TxTypeCreateAccountDepositTransfer, TxTypeDepositTransfer, TxTypeForceTransfer,
		TxTypeForceExit, TxTypeTransferToEthAddr, TxTypeTransferToBJJ:
		return &txTypeCasted, nil
	default:
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, %s is not a valid option. Check the valid options in the documentation",
			"type", txType,
		))
	}
}

// StringToL2TxState converts string to l2 transaction state
func StringToL2TxState(txState string) (*PoolL2TxState, error) {
	if txState == "" {
		return nil, nil
	}
	txStateCasted := PoolL2TxState(txState)
	switch txStateCasted {
	case PoolL2TxStatePending, PoolL2TxStateForged, PoolL2TxStateForging, PoolL2TxStateInvalid:
		return &txStateCasted, nil
	default:
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, %s is not a valid option. Check the valid options in the documentation",
			"state", txState,
		))
	}
}

// QueryAccount is a representation of an account with accountIndex and its token symbol
type QueryAccount struct {
	AccountIndex *Idx
	Symbol       string
}

// StringToIdx converts string to account index
func StringToIdx(idxStr, name string) (QueryAccount, error) {
	if idxStr == "" {
		return QueryAccount{}, nil
	}
	splitted := strings.Split(idxStr, ":")
	const expectedLen = 3
	if len(splitted) != expectedLen || splitted[0] != "cb" {
		return QueryAccount{}, tracerr.Wrap(fmt.Errorf(
			"invalid %s, must follow this: cb:<tokenSymbol>:index", name))
	}
	idxInt, err := strconv.Atoi(splitted[2])
	idx := Idx(idxInt)
	return QueryAccount{AccountIndex: &idx, Symbol: splitted[1]}, tracerr.Wrap(err)
}

// CbStringToEthAddr converts cb ethereum address to ethereum address
func CbStringToEthAddr(addrStr, name string) (*ethCommon.Address, error) {
	if addrStr == "" {
		return nil, nil
	}
	splitted := strings.Split(addrStr, "cb:")
	if len(splitted) != 2 || len(splitted[1]) != 42 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^cb:0x[a-fA-F0-9]{40}$", name))
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(splitted[1]))
	return &addr, tracerr.Wrap(err)
}

// CbStringToBJJ converts cb ethereum address string to bjj
func CbStringToBJJ(bjjStr, name string) (*babyjub.PublicKeyComp, error) {
	if bjjStr == "" {
		return nil, nil
	}
	const decodedLen = 33
	splitted := strings.Split(bjjStr, "cb:")
	if len(splitted) != 2 || len(splitted[1]) != 44 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^cb:[A-Za-z0-9+/=]{44}$",
			name))
	}
	decoded, err := base64.RawURLEncoding.DecodeString(splitted[1])
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, error decoding base64 string: %s",
			name, err.Error()))
	}
	if len(decoded) != decodedLen {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, error decoding base64 string: unexpected byte array length",
			name))
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return nil, tracerr.Wrap(fmt.Errorf("invalid %s, checksum failed",
			name))
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return &bjjComp, nil
}

// StringToEthAddr converts string to ethereum address
func StringToEthAddr(ethAddrStr string) (*ethCommon.Address, error) {
	if ethAddrStr == "" {
		return nil, nil
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(ethAddrStr))
	return &addr, tracerr.Wrap(err)
}

// BjjToString converts baby jub jub public key to string
func BjjToString(bjj babyjub.PublicKeyComp) string {
	pkComp := [32]byte(bjj)
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "cb:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

// EthAddrToCb converts ethereum address to chainbing ethereum address
func EthAddrToCb(addr ethCommon.Address) string {
	return "cb:" + addr.String()
}

// IdxToCb converts account index to cb account index with token symbol
func IdxToCb(idx Idx, tokenSymbol string) string {
	return "cb:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

// StrCbIdx is used to unmarshal CbIdx directly into an alias of Idx
type StrCbIdx struct {
	Idx         Idx
	TokenSymbol string
}

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
	*s = StrCbIdx{
		Idx:         Idx(idxInt),
		TokenSymbol: splitted[0],
	}
	return nil
}

// StrCbEthAddr is used to unmarshal CbEthAddr directly into an alias of ethCommon.Address
type StrCbEthAddr ethCommon.Address

// UnmarshalText unmarshals a StrCbEthAddr
func (s *StrCbEthAddr) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = StrCbEthAddr(EmptyAddr)
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

// StrCbBJJ is used to unmarshal CbBJJ directly into an alias of babyjub.PublicKeyComp
type StrCbBJJ babyjub.PublicKeyComp

// UnmarshalText unmarshalls a StrCbBJJ
func (s *StrCbBJJ) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = StrCbBJJ(EmptyBJJComp)
		return nil
	}
	bjj, err := CbStringToBJJ(string(text), "")
	if err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrCbBJJ(*bjj)
	return nil
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
