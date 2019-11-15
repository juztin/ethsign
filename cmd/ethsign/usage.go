package main

const USAGE = `
ARGUMENTS
  --abi f͟i͟l͟e͟
          Contract Application Binary Interface file.

  --bin f͟i͟l͟e͟
          Contract compiled bytecode file.

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
      [DEFAULT 1]

  --gasLimit n͟
          The maximum amount of gas the transaction may consume.
      [DEFAULT 100000]

  --key f͟i͟l͟e͟
          File containing either the raw private key, or a Go-Ethereum keystore file.
          A password prompt will occur for keystore files.
      [REQUIRED]

  --nonce n͟
          The next nonce of the address for the --key file.
      [DEFAULT 0]

  --to a͟d͟d͟r͟e͟s͟s͟
          The recipient of either the ether, the contract address of the invocation, or both.
      Not required when signing a transaction for contract deployment.

  --value n͟
          The amount in Ether to send with the transaction.
      [DEFAULT 0]

EXAMPLES

  Sending ether:
    ethsign ether --to 0x0000000000000000000000000000000000000000 --key keyfile.json --value 0.05

  Transfer ERC-20 tokens:
    ethsign call "transfer(address,uint256)" 0xffffffffffffffffffffffffffffffffffffffff 42 --key keyfile.json

  Function call from contract ABI
    ethsign call funcName arg1 arg2 --to 0x0000000000000000000000000000000000000000 --abi contract.abi --key keyfile.txt

  Contract deployment, with constructor arguments
    ethsign deploy arg1 arg2 --abi contract.abi --bin contract.bin --key keyfile.json
    ethsign deploy constructor(string,uint256) arg1 arg2 --bin contract.bin --key keyfile.json

  Generating a QR-Code (using 'qr-code' tool: 'go get github.com/juztin/qr-code')
    qrcode echo "https://etherscan.io/pushTx?hex=0x$(ethsign ether --to 0xffffffffffffffffffffffffffffffffffffffff --value 0.25 --key keyfile.key --nonce 42 -gasPrice 2 -gasLimit 21000)" > transaction.png
`
