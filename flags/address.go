package flags

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type AddressFlag struct {
	Value common.Address
}

func (f *AddressFlag) String() string {
	return f.Value.String()
}

func (f *AddressFlag) Set(value string) error {
	if !common.IsHexAddress(value) {
		return errors.New("Invalid address for value")
	}
	f.Value = common.HexToAddress(value)
	return nil
}

func (f *AddressFlag) IsSet() bool {
	for _, b := range f.Value {
		if b != 0 {
			return true
		}
	}
	return false
}

func Address(value common.Address) AddressFlag {
	return AddressFlag{value}
}
