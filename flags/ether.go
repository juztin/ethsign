package flags

import (
	"errors"
	"math/big"
)

var ether = big.NewFloat(1000000000000000000)

// EtherFlag is a flag to convert from ether to wei
type EtherFlag struct {
	i *big.Int
}

// String returns the string value
func (f *EtherFlag) String() string {
	return f.i.String()
}

// Set converts the given value from ether to wei, and sets it
func (f *EtherFlag) Set(value string) error {
	amount, ok := new(big.Float).SetString(value)
	if !ok {
		return errors.New("Invalid number for big.Float")
	}
	amount.Mul(amount, ether).Int(f.i)
	return nil
}

// Value returns the value
func (f *EtherFlag) Value() *big.Int {
	return f.i
}

// Ether returns a new EtherFlag set to the given value, in wei
func Ether(value *big.Int) EtherFlag {
	return EtherFlag{value}
}
