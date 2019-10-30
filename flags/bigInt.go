package flags

import (
	"errors"
	"math/big"
)

type BigIntFlag struct {
	i *big.Int
}

func (f *BigIntFlag) String() string {
	return f.i.String()
}

func (f *BigIntFlag) Set(value string) error {
	i, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return errors.New("Invalid number for big.Int")
	}
	f.i = i
	return nil
}

func (f *BigIntFlag) Value() *big.Int {
	return f.i
}

func BigInt(value *big.Int) BigIntFlag {
	return BigIntFlag{value}
}
