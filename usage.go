package ethsign

const USAGE = `
ARGUMENTS
  --abi f͟i͟l͟e͟
          Contract Application Binary Interface file.

  --bin f͟i͟l͟e͟
          Contract compiled bytecode file.
		  Requires that the 'funcName' argument must be blank, eg.

		    ethsign "" --key keyfile.json --abi contract.abi --bin contract.bin --nonce 5

  --chain n͟
          The Ethereum chain, for EIP-155 signing.
              1 - mainnet
		      2 - Morden
		      3 - Ropsten
		      4 - Rinkby
		      5 - Goerli
		      6 - Kovan
		   1337 - Geth private chain
		   [default 1337]

  --gasPrice n͟
          The gas price in Gwei.
		  [default 1]

  --gasLimit n͟
          The maximum amount of gas the transaction may consume.
		  [default 21000]

  --key f͟i͟l͟e͟
          File containing either the raw private key, or a Go-Ethereum keystore file.
          A password prompt will occur for keystore files.
		  [REQUIRED]

  --nonce n͟
          The next nonce of the address for the --key file.
		  [REQUIRED]

  --to a͟d͟d͟r͟e͟s͟s͟
          The recipient of either the ether, the contract address of the invocation, or both.
		  Not required when signing a transaction for contract deployment.

  --value n͟
          The amount in Ether to send with the transaction.
		  [default 0]

EXAMPLES

  ethsign --to 0x0000000000000000000000000000000000000000 --key keyfile.json --abi contract.abi --nonce 5 --gasPrice 2 --value 0.05 --chain 1
  ethsign funcName arg1 arg2... --to 0x0000000000000000000000000000000000000000 --key keyfile.txt --abi contract.abi --nonce 5 --gasPrice 2 --gasLimit 100000 --value 0 --chain 1
  ethsign "" arg1 arg2... --key keyfile.json --abi contract.abi --nonce 5 --gasPrice 2 --gasLimit 100000 --value 0 --chain 1
`
