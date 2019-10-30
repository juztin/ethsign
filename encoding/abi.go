package encoding

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func DecodeArgs(args abi.Arguments, s ...string) ([]interface{}, error) {
	if len(args) != len(s) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(args), len(s))
	}
	var err error
	packed := make([]interface{}, len(s), len(s))
	for i := range args {
		packed[i], err = convertTo(args[i].Type, s[i])
		if err != nil {
			break
		}
	}
	return packed, err
}

var addressType = reflect.TypeOf(common.Address{})
var bigIntType = reflect.TypeOf((*big.Int)(nil))
var boolType = reflect.TypeOf(false)
var uint8Type = reflect.TypeOf(uint8(0))
var uint16Type = reflect.TypeOf(uint16(0))
var uint32Type = reflect.TypeOf(uint32(0))
var uint64Type = reflect.TypeOf(uint64(0))
var int8Type = reflect.TypeOf(int8(0))
var int16Type = reflect.TypeOf(int16(0))
var int32Type = reflect.TypeOf(int32(0))
var int64Type = reflect.TypeOf(int64(0))

func convertTo(t abi.Type, s string) (interface{}, error) {
	//fmt.Printf("###### kind: %s, type: %s, string: %s\n", t.Kind, t.Type, t.String())
	//
	// TODO: Check if it's easier to switch on `t.String()`
	//

	var err error
	var o interface{} = s

	switch t.Type {
	case nil:
		break
	case boolType:
		o, err = strconv.ParseBool(s)
	case int8Type:
		o, err = strconv.ParseInt(s, 10, 8)
		if err == nil {
			o = int8(o.(int64))
		}
	case int16Type:
		o, err = strconv.ParseInt(s, 10, 16)
		if err == nil {
			o = int16(o.(int64))
		}
	case int32Type:
		o, err = strconv.ParseInt(s, 10, 32)
		if err == nil {
			o = int32(o.(int64))
		}
	case int64Type:
		o, err = strconv.ParseInt(s, 10, 64)
		if err == nil {
			o = o.(int64)
		}
	case uint8Type:
		o, err = strconv.ParseUint(s, 10, 8)
		if err == nil {
			o = uint8(o.(uint64))
		}
	case uint16Type:
		o, err = strconv.ParseUint(s, 10, 16)
		if err == nil {
			o = uint16(o.(uint64))
		}
	case uint32Type:
		o, err = strconv.ParseUint(s, 10, 32)
		if err == nil {
			o = uint32(o.(uint64))
		}
	case uint64Type:
		o, err = strconv.ParseUint(s, 10, 64)
		if err == nil {
			o = o.(uint64)
		}
	case bigIntType:
		var ok bool
		o, ok = new(big.Int).SetString(s, 10)
		if !ok {
			err = errors.New("Invalid big.Int")
		}
		break
	case addressType:
		if !common.IsHexAddress(s) {
			err = errors.New("Invalid address")
		}
		o = common.HexToAddress(s)
		break
	}

	return o, err
}
