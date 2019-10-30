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

	"github.com/juztin/ethsign"
	"github.com/juztin/ethsign/encoding"
	"github.com/juztin/ethsign/flags"
)

var (
	// args
	args     []string
	funcName string

	// flags
	abiFlag       flags.FileFlag
	binFlag       flags.FileFlag
	keyFlag       flags.FileFlag
	keystoreFlag  flags.FileFlag
	recipientFlag flags.AddressFlag

	chainFlag    = flags.BigInt(big.NewInt(1337))
	helpFlag     = flag.Bool("help", false, "Print ethsign usage")
	nonceFlag    = flag.Uint64("nonce", 0, "Next nonce for the address signing the transaction")
	gasPriceFlag = flags.BigInt(big.NewInt(21000))
	gasLimitFlag = flag.Uint64("gasLimit", 1000000, "The gas limit, in Gwei")
	valueFlag    = flags.BigInt(big.NewInt(0))
)

func init() {
	//flag.IntVar(chain, "c", "", "Chain ID")
	//flag.StringVar(keyFile, "k", "", "Private key file path")

	flag.Var(&abiFlag, "abi", "Contract ABI file")
	flag.Var(&binFlag, "bin", "Contract BIN file")
	flag.Var(&chainFlag, "chain", "Chain ID (default l337)")
	flag.Var(&keyFlag, "key", "Private key filepath")
	flag.Var(&keystoreFlag, "keystore", "Private go-ethereum keystore filepath")
	flag.Var(&recipientFlag, "to", "The recipient address to send the transaction to")
	flag.Var(&gasPriceFlag, "gasPrice", "The gas price to use, in Gwei (default 21000)")
	flag.Var(&valueFlag, "value", "The amount of Ether to send with the transaction (default 0)")

	pos := 1
	for ; pos < len(os.Args); pos++ {
		if len(os.Args[pos]) > 0 && os.Args[pos][0] == '-' {
			break
		}
	}
	//flag.Usage = usage
	//flag.Parse()
	if pos > 1 {
		funcName = os.Args[1]
		if pos > 2 {
			args = os.Args[2:pos]
		}
	}
	flag.CommandLine.Parse(os.Args[pos:])
}

func readABI(abiFile string) (abi.ABI, error) {
	// Read ABI file
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

//func readABIandBIN(abiFile, binFile string) (abi.ABI, []byte, error) {
//	a, err := readABI(abiFile)
//	if err != nil {
//		return a, nil, err
//	}
//	b, err := ioutil.ReadFile(binFile)
//	return a, b, err
//}

func deployInput(args []string, abiFile, binFile string) ([]byte, error) {
	//var err error
	//if abiFlag.String() == "" {
	//	err = errors.New("flag -abi is required when -bin is supplied")
	//} else if recipientFlag.IsSet() {
	//	err = errors.New("flag -to can't be set for contract deployments (when -bin is set)")
	//}
	//if err != nil {
	//	fmt.Println(err)
	//	flag.Usage()
	//	os.Exit(1)
	//}

	input, err := callInput("constructor", args, abiFile)
	if err != nil {
		return nil, err
	}
	b, err := readBin(binFile)
	return append(b, input...), err
}

func sendEtherInput() ([]byte, error) {
	return nil, nil
}

func callInput(funcName string, args []string, abiFile string) ([]byte, error) {
	a, err := readABI(abiFile)
	if err != nil {
		return nil, err
	}

	// Ensure the given function name exists within the ABI
	var m abi.Method
	ok := true
	if funcName == "constructor" {
		funcName = ""
		m = a.Constructor
	} else if m, ok = a.Methods[funcName]; !ok {
		return nil, errors.New("missing function in ABI")
	}

	// Convert args to matching types
	funcArgs, err := encoding.DecodeArgs(m.Inputs, args...)
	if err != nil {
		return nil, err
	}

	// Generate packed call
	input, err := a.Pack(funcName, funcArgs...)
	return input, err
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
	if *helpFlag {
		fmt.Println(ethsign.USAGE)
		os.Exit(0)
	}

	var err error
	var input []byte
	var tx *types.Transaction

	if len(args) == 0 {
		input, err = sendEtherInput()
		tx = types.NewTransaction(*nonceFlag, recipientFlag.Value, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), input)
	} else if binFlag.String() != "" {
		input, err = deployInput(args, abiFlag.String(), binFlag.String())
		err = errors.New("boom")
		tx = types.NewContractCreation(*nonceFlag, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), input)
	} else {
		input, err = callInput(funcName, args, abiFlag.String())
		tx = types.NewTransaction(*nonceFlag, recipientFlag.Value, valueFlag.Value(), *gasLimitFlag, gasPriceFlag.Value(), input)
	}

	if keystoreFlag.String() != "" {
		tx, err = signTxWithKeystore(tx, chainFlag.Value(), keystoreFlag.String())
	} else {
		tx, err = signTxWithKey(tx, chainFlag.Value(), keyFlag.String())
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t := types.Transactions{tx}
	rawTxBytes := t.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	fmt.Printf(rawTxHex)

	//fmt.Printf("input: 0x%x\n", input)
	//fmt.Println("error:", err)
	//fmt.Println("----------------------------------------")
	//fmt.Println("args:", args)
	//fmt.Println("function: ", funcName)

	//fmt.Println("abi:", abiFlag.String())
	//fmt.Println("bin:", binFlag.String())
	//fmt.Println("chain:", chainFlag.Value())
	//fmt.Println("keyFile:", keyFlag.String())
	//fmt.Println("nonce:", *nonceFlag)
	//fmt.Println("gasPrice:", gasPriceFlag.Value())
	//fmt.Println("gasLimit:", *gasLimitFlag)
	//fmt.Println("recipient:", recipientFlag.String())
	//fmt.Println("value:", valueFlag.Value())

	//fmt.Println("signedTx:", rawTxHex)
}
