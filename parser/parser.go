package parser

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ParseMethodString(s string) ([]byte, []string, error) {
	// Must be at-least "f()"
	if len(s) <= 3 {
		return nil, nil, errors.New("Invalid, or empty call")
	}
	// Remove all whitespace – " test( string, bool)" => "test(string,bool)"
	s = strings.Replace(s, " ", "", -1)
	// Ensure we end with a paren
	if s[len(s)-1] != ')' {
		return nil, nil, errors.New("invalid method signature")
	}
	// Grab the first paren
	i := strings.Index(s, "(")
	// Must be after the first character "f()", "ffff()", etc.
	if i < 1 {
		return nil, nil, errors.New("invalid method signature")
	}
	args := strings.Split(s[i+1:len(s)-1], ",")
	sig := crypto.Keccak256([]byte(s))[:4]
	return sig, args, nil
}

func ParseMethodArgs(types, args []string) ([]byte, error) {
	if len(types) != len(args) {
		return nil, fmt.Errorf("Mismatched length, expected %d got %d", len(types), len(args))
	}
	var b []byte
	var err error
	for i := range types {
		// Get parsed value
		o, err := ParseValue(types[i], args[i])
		if err != nil {
			break
		}
		// Pack the variable
		t, err := abi.NewType(types[i], types[i], nil)
		if err != nil {
			break
		}
		a := abi.Arguments{{Type: t}}
		p, err := a.Pack(o)
		if err != nil {
			break
		}
		b = append(b, p...)
	}
	return b, err
}

func ParseMethod(methodSig string, args []string) ([]byte, error) {
	data, methodArgs, _ := ParseMethodString(methodSig)
	input, err := ParseMethodArgs(methodArgs, args)
	if err != nil {
		return data, err
	}
	return append(data, input...), nil
}

func ParseValue(s, v string) (interface{}, error) {
	var o interface{}
	switch s {
	case "address":
		if common.IsHexAddress(v) == false {
			return o, fmt.Errorf("Invalid address '%s'", v)
		}
		o = common.HexToAddress(v)
		break
	case "bool":
		t, err := strconv.ParseBool(v)
		if err != nil {
			return o, fmt.Errorf("Invalid bool '%s': %w", t, err)
		}
		o = t
		break
	case "string":
		o = v
		break
	case "bytes1", "bytes2", "bytes3", "bytes4", "bytes5", "bytes6", "bytes7", "bytes8",
		"bytes9", "bytes10", "bytes11", "bytes12", "bytes13", "bytes14", "bytes15", "bytes16",
		"bytes17", "bytes18", "bytes19", "bytes20", "bytes21", "bytes22", "bytes23", "bytes24",
		"bytes25", "bytes26", "bytes27", "bytes28", "bytes29", "bytes30", "bytes31", "bytes32":
		h, err := hex.DecodeString(v)
		if err != nil {
			return o, fmt.Errorf("Invalid bytes '%s': %w", h, err)
		}
		size, _ := strconv.ParseInt(s[5:], 10, 8)
		if len(h) != int(size) {
			return o, fmt.Errorf("Invalid bytes length, expected %d got %d", s[5:], len(h))
		}
		o = h
		break
	case "uint", "int",
		"uint256", "int256",
		"uint248", "int248",
		"uint240", "int240",
		"uint232", "int232",
		"uint224", "int224",
		"uint216", "int216",
		"uint208", "int208",
		"uint200", "int200",
		"uint192", "int192",
		"uint184", "int184",
		"uint176", "int176",
		"uint168", "int168",
		"uint160", "int160",
		"uint152", "int152",
		"uint144", "int144",
		"uint136", "int136",
		"uint128", "int128",
		"uint120", "int120",
		"uint112", "int112",
		"uint104", "int104",
		"uint96", "int96",
		"uint88", "int88",
		"uint80", "int80",
		"uint72", "int72":
		n, ok := new(big.Int).SetString(v, 10)
		if !ok {
			return o, fmt.Errorf("Invalid number '%s'", v)
		}
		o = n
		break
	case "uint64", "uint56", "uint48", "uint40":
		d, err := parseUint(v, s[4:], 8)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = uint64(d)
		break
	case "uint32", "uint24":
		d, err := parseUint(v, s[4:], 8)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = uint32(d)
		break
	case "uint16":
		d, err := parseUint(v, s[4:], 8)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = uint16(d)
		break
	case "uint8":
		d, err := parseUint(v, s[4:], 8)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = uint8(d)
		break
	case "int64", "int56", "int48", "int40":
		d, err := parseInt(v, s[4:], 64)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = int64(d)
		break
	case "int32", "int24":
		d, err := parseInt(v, s[4:], 32)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = int32(d)
		break
	case "int16":
		d, err := parseInt(v, s[4:], 16)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = int16(d)
		break
	case "int8":
		d, err := parseInt(v, s[4:], 8)
		if err != nil {
			return o, fmt.Errorf("Invalid number '%s': %w", v, err)
		}
		o = int8(d)
		break
	default:
		return o, fmt.Errorf("Invalid type '%s'", s)
		break
	}
	return o, nil
}

func parseInt(value, base string, bitSize int) (int64, error) {
	b, err := strconv.ParseInt(base, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, int(b))
}

func parseUint(value, base string, bitSize int) (uint64, error) {
	b, err := strconv.ParseInt(base, 10, bitSize)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(value, 10, int(b))
}
