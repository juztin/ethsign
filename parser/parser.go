package parser

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var defaults = map[string]string{
	"address": "0x0000000000000000000000000000000000000000",
	"bool":    "false",
	"string":  "",
}

// ParseValue parses the given value to the corresponding kind
func ParseValue(kind, value string) (interface{}, error) {
	begin := strings.Index(kind, "[")
	// Non array
	if begin < 0 {
		return parseValue(kind, value)
	}
	// Array
	k := kind[:begin]
	args := kind[begin:]
	var ok bool
	ok, value = isValidArrayArgs(args, value)
	if !ok {
		return nil, errors.New("Invalid or mismatched array signature and/or value")
	}
	t, err := typeForKind(k)
	if err != nil {
		return nil, err
	}
	o, err := parseArray(t, k, value)
	return o, err
}

// ParseMethod parses the given args to the corresponding types found within the method signature, returning the raw data
func ParseMethod(method string, args []string) ([]byte, error) {
	// Remove all whitespace – " test( string, bool)" => "test(string,bool)"
	method = strings.Replace(method, " ", "", -1)
	sig, methodArgs, err := parseMethodString(method)
	if err != nil {
		return nil, err
	}
	data, err := parseMethodArgs(methodArgs, args)
	if err != nil {
		return data, err
	}
	return append(sig, data...), nil
}

func parseMethodString(s string) ([]byte, []string, error) {
	// Must have an open parent and be at-least "f()"
	start, end := strings.Index(s, "("), strings.Index(s, ")")
	if start < 1 || end < start || end != len(s)-1 || len(s) <= 3 {
		return nil, nil, errors.New("Invalid call")
	}
	args := strings.Split(s[start+1:end], ",")
	sig := crypto.Keccak256([]byte(s))[:4]
	return sig, args, nil
}

func parseMethodArgs(types, args []string) ([]byte, error) {
	if len(types) != len(args) {
		return nil, fmt.Errorf("Mismatched length, expected %d got %d", len(types), len(args))
	}
	var b []byte
	for i := range types {
		// Get parsed value
		o, err := ParseValue(types[i], args[i])
		if err != nil {
			return b, err
		}
		// Pack the variable
		t, err := abi.NewType(types[i], types[i], nil)
		if err != nil {
			return b, err
		}
		a := abi.Arguments{{Type: t}}
		p, err := a.Pack(o)
		if err != nil {
			return b, err
		}
		b = append(b, p...)
	}
	return b, nil
}

func isValidArrayArgs(signature, value string) (bool, string) {
	// Very basic validation check ensuring signature and value end with a closing bracket, and the value also starts with an opening bracket
	if len(signature) < 2 || len(value) < 2 {
		return false, ""
	} else if signature[len(signature)-1] != ']' {
		return false, ""
	} else if value[0] != '[' || value[len(value)-1] != ']' {
		return false, ""
	}

	// Remove all non-quotes spaces (TODO: do we really need to check all other chars? tabs, etc.)
	// Ensure '[' have matching ']'
	b := true
	start, end := 0, 0
	v := make([]byte, len(value))
	count := 0
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '"':
			b = !b
			continue
		case ' ':
			if b {
				continue
			}
		case '[':
			start++
			break
		case ']':
			end++
			break
		}
		if start < 0 || end < 0 {
			return false, ""
		}
		v[count] = value[i]
		count++
	}
	b = b && start-end == 0
	return b, string(v[:count])
}

func typeForKind(kind string) (reflect.Type, error) {
	// Get the type, using the default value
	d, ok := defaults[kind]
	if !ok {
		d = "0"
	}
	v, err := parseValue(kind, d)
	if err != nil {
		return nil, err
	}
	return reflect.TypeOf(v), nil
}

func parseArray(t reflect.Type, kind, val string) (interface{}, error) {
	// TODO: Support multi-dimensional arrays
	// Create the array and populate it with the parsed valued
	o := reflect.New(reflect.SliceOf(t)).Elem()
	s := strings.Split(val[1:len(val)-1], ",")
	for i := range s {
		p, err := parseValue(kind, s[i])
		if err != nil {
			return o, err
		}
		o = reflect.Append(o, reflect.ValueOf(p))
	}
	return o.Interface(), nil
}

func parseValue(s, v string) (interface{}, error) {
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
