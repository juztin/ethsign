package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/juztin/ethsign/encoding"
	"github.com/juztin/ethsign/flags"
	"github.com/juztin/ethsign/parser"
)

type command int
type signFunc func(*types.Transaction, *big.Int, string) (*types.Transaction, error)

const (
	CALL = iota
	DEPLOY
	ETHER
)

var (
	// args
	args       []string
	cmd        command
	keyPath    string
	method     string
	methodArgs []string
	signFn     signFunc

	// flags
	abiFlag       flags.FileFlag
	binFlag       flags.FileFlag
	keyFlag       flags.FileFlag
	keystoreFlag  flags.FileFlag
	recipientFlag flags.AddressFlag

	chainFlag    = flags.BigInt(big.NewInt(1337))
	gasPriceFlag = flags.Ether(big.NewInt(1), flags.GWEI)
	gasLimitFlag = flag.Uint64("gasLimit", 100000, "The gas limit, in Gwei")
	helpFlag     = flag.Bool("help", false, "Print ethsign usage")
	nonceFlag    = flag.Uint64("nonce", 0, "Next nonce for the address signing the transaction")
	valueFlag    = flags.Ether(big.NewInt(0), flags.ETHER)
)

func init() {
	//flag.IntVar(chain, "c", "", "Chain ID")
	//flag.StringVar(keyFile, "k", "", "Private key file path")

	flag.Var(&abiFlag, "abi", "Contract ABI file")
	flag.Var(&binFlag, "bin", "Contract BIN file, for contract deployments")
	flag.Var(&chainFlag, "chain", "Chain ID (default l337)")
	flag.Var(&gasPriceFlag, "gasPrice", "The gas price to use, in Gwei (default 1)")
	flag.Var(&keyFlag, "key", "Private key filepath")
	flag.Var(&keystoreFlag, "keystore", "Private go-ethereum keystore filepath")
	flag.Var(&recipientFlag, "to", "The recipient address to send the transaction to")
	flag.Var(&valueFlag, "value", "The amount of Ether to send with the transaction (default 0)")

	//pos := 0
	//for i := 1; i < len(os.Args); i++ {
	//	if pos == 0 {
	//		if len(os.Args[i]) > 0 && os.Args[i][0] == '-' {
	//			pos = i
	//		}
	//	} else if os.Args[i][0] != '-' {
	//		checkErr(errors.New("Argument exists after flag(s)"))
	//	}
	//}

	pos := 1
	for ; pos < len(os.Args); pos++ {
		if len(os.Args[pos]) > 0 && os.Args[pos][0] == '-' {
			break
		}
	}
	flag.Usage = usage
	flag.Parse()
	if pos < 2 {
		checkErr(errors.New("Missing required command: [ether, call, deploy]"))
	}
	switch os.Args[1] {
	case "call":
		cmd = CALL
		break
	case "deploy":
		cmd = DEPLOY
		break
	case "ether":
		cmd = ETHER
		break
	case "help":
		flag.Usage()
		break
	default:
		checkErr(fmt.Errorf("Invalid command: '%s', must be one of [ether, call, deploy]", os.Args[1]))
	}
	if pos > 2 {
		args = os.Args[2:pos]
	}
	flag.CommandLine.Parse(os.Args[pos:])
}

func checkErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func usage() {
	fmt.Println(USAGE)
}

func validateArgs() error {
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Validate key/keystore
	keyPath = keyFlag.String()
	if keyPath == "" {
		return errors.New("Must specify key, or keystore file [--key]")
	}
	fi, err := os.Stat(keyPath)
	if err != nil {
		return fmt.Errorf("Must specify a valid key, or keystore file, %v", err)
	}
	if fi.Size() == 0x40 {
		signFn = signTxWithKey
	} else {
		signFn = signTxWithKeystore
	}

	// Ensure `Value` is non-negative
	if valueFlag.Value().Cmp(big.NewInt(0)) < 0 {
		return errors.New("Can't send negative Ether")
	}

	// Ensure `Nonce` is non-negative
	if *nonceFlag < 0 {
		return errors.New("Nonce must bea non-negative")
	}

	switch cmd {
	case CALL:
		if len(args) == 0 {
			if abiFlag.String() == "" {
				err = errors.New("Must specify function signature")
			} else {
				err = errors.New("Must specify ABI function name")
			}
		}
		method = args[0]
		methodArgs = args[1:]
		break
	case DEPLOY:
		if binFlag.String() == "" {
			err = errors.New("Must specify bin file for contract deployment [--bin]")
		} else if recipientFlag.IsSet() {
			err = errors.New("Recipient can't be set for contract deployment")
		}
		if len(args) == 0 {
			method = "constructor()"
		} else {
			method = args[0]
			methodArgs = args[1:]
		}
		break
	case ETHER:
		if recipientFlag.IsSet() == false {
			err = errors.New("Must specify a valid recipient to send ether [--recipient]")
		} else if len(args) > 1 {
			err = errors.New("Can't supply multiple arguments/messages within a transaction")
		}
		break
	}
	return err
}

func readABI(abiFile string) (abi.ABI, error) {
	r, err := os.Open(abiFile)
	if err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(r)
}

func readBin(binFile string) ([]byte, error) {
	b, err := ioutil.ReadFile(binFile)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(b))
}

func callInputABI(methodSig string, args []string, abiFile string) ([]byte, error) {
	a, err := readABI(abiFile)
	if err != nil {
		return nil, err
	}

	// Ensure the given function name exists within the ABI
	var m abi.Method
	ok := true
	if methodSig == "" {
		m = a.Constructor
	} else if m, ok = a.Methods[methodSig]; !ok {
		return nil, errors.New("missing function in ABI")
	}

	// Convert args to matching types
	funcArgs, err := encoding.DecodeArgs(m.Inputs, args...)
	if err != nil {
		return nil, err
	}

	// Generate packed call
	input, err := a.Pack(methodSig, funcArgs...)
	return input, err
}

func deployInputABI(methodSig string, args []string, binFile, abiFile string) ([]byte, error) {
	input, err := callInputABI("", args, abiFile)
	if err != nil {
		return input, err
	}
	b, err := readBin(binFile)
	return append(b, input...), err
}

func deployInputRaw(methodSig string, args []string, binFile string) ([]byte, error) {
	input, err := parser.ParseConstructor(methodSig, args)
	if err != nil {
		return input, err
	}
	b, err := readBin(binFile)
	return append(b, input...), err
}

func etherInput(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, nil
	}
	return []byte(args[0]), nil
}

func signTxWithKeystore(tx *types.Transaction, chainID *big.Int, keyPath string) (*types.Transaction, error) {
	// Read keystore file
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// Prompt for key password
	fmt.Fprint(os.Stderr, "Passphrase: ")
	p, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}

	// Decrypt and sign
	k, err := keystore.DecryptKey(b, string(p))
	return types.SignTx(tx, types.NewEIP155Signer(chainID), k.PrivateKey)
}

func signTxWithKey(tx *types.Transaction, chainID *big.Int, keyPath string) (*types.Transaction, error) {
	// Read keystore file
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// Convert to ECDSA
	k, err := crypto.HexToECDSA(string(b))
	if err != nil {
		return nil, err
	}
	return types.SignTx(tx, types.NewEIP155Signer(chainID), k)
}

func main() {
	err := validateArgs()
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%s\n\nfor help, '%s --help'", err, os.Args[0])
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Create transaction
	var data []byte
	var tx *types.Transaction
	switch cmd {
	case CALL:
		if abiFlag.String() == "" {
			data, err = parser.ParseMethod(method, methodArgs)
		} else {
			data, err = callInputABI(method, methodArgs, abiFlag.String())
		}
		tx = types.NewTransaction(*nonceFlag, recipientFlag.Value, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), data)
		break
	case DEPLOY:
		if abiFlag.String() == "" {
			data, err = deployInputRaw(method, methodArgs, binFlag.String())
		} else {
			data, err = deployInputABI(method, methodArgs, binFlag.String(), abiFlag.String())
		}
		tx = types.NewContractCreation(*nonceFlag, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), data)
		break
	case ETHER:
		data, err = etherInput(args)
		tx = types.NewTransaction(*nonceFlag, recipientFlag.Value, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), data)
		break
	}

	// Sign transaction
	checkErr(err)
	tx, err = signFn(tx, chainFlag.Value(), keyPath)
	checkErr(err)

	// Print raw, signed, hex-string transaction
	t := types.Transactions{tx}
	rawTx := t.GetRlp(0)
	fmt.Printf("0x%x", rawTx)
}
