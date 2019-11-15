package encoding

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/juztin/ethsign/parser"
)

func DecodeArgs(args abi.Arguments, s ...string) ([]interface{}, error) {
	if len(args) != len(s) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(args), len(s))
	}
	var err error
	packed := make([]interface{}, len(s), len(s))
	for i := range args {
		packed[i], err = parser.ParseValue(args[i].Type.String(), s[i])
		if err != nil {
			break
		}
	}
	return packed, err
}
