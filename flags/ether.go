package flags

import (
	"errors"
	"math/big"
)

var ether = big.NewFloat(1000000000000000000)

type Unit int64

const (
	WEI    Unit = 1
	KWEI        = 1000
	MWEI        = 1000000
	GWEI        = 1000000000
	SZABO       = 1000000000000
	FINNEY      = 1000000000000000
	ETHER       = 1000000000000000000
	KETHER      = 1000000000000000000000
	METHER      = 1000000000000000000000000
	GETHER      = 1000000000000000000000000000
	TETHER      = 1000000000000000000000000000000
)

// EtherFlag is a flag to convert from ether to wei
type EtherFlag struct {
	unit  Unit
	value *big.Int
}

// String returns the string value
func (f *EtherFlag) String() string {
	return f.value.String()
}

// Set converts the given value from ether to wei, and sets it
func (f *EtherFlag) Set(value string) error {
	amount, ok := new(big.Float).SetString(value)
	if !ok {
		return errors.New("Invalid number for big.Float")
	}
	u := new(big.Float).SetInt64(int64(f.unit))
	amount.Mul(amount, u).Int(f.value)
	return nil
}

// Value returns the value
func (f *EtherFlag) Value() *big.Int {
	return f.value
}

// Ether returns a new EtherFlag set to the given value, in wei
func Ether(value *big.Int, u Unit) EtherFlag {
	return EtherFlag{u, value}
}
