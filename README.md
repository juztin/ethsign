# Ethereum Signed Transaction Generator

Utility to generate raw/signed Ethereum transactions offline.

### Install

```shell
% go install github.com/juztin/ethersign/cmd/ethersign
```

### Usage

##### Sending Ethereum

```
ethsign send --to 0x0000000000000000000000000000000000000000 --key keyfile.json --value 0.05
```

##### Send a message to a contract _(ERC-20 transfer)_

**without ABI**
```
ethsign call "transfer(address,uint256)" 0xffffffffffffffffffffffffffffffffffffffff 42 --key keyfile.json
```

**with ABI**
```
ethsign call funcName arg1 arg2 --to 0x0000000000000000000000000000000000000000 --abi contract.abi --key keyfile.txt
```

##### Contract Deployment

**without ABI**
```
ethsign deploy --abi contract.abi --bin contract.bin --key keyfile.json
ethsign deploy arg1 arg2 --abi contract.abi --bin contract.bin --key keyfile.json
```
**with ABI**
```
ethsign deploy --bin contract.bin --key keyfile.json
ethsign deploy constructor(string,uint256) arg1 arg2 --bin contract.bin --key keyfile.json
```


##### Generating a QR Code

Install [qr-code](https://github.com/juztin/qr-code) _(or use any other qr-code generator)_
```
go get github.com/juztin/qr-code
```

Create a QR Code for submission directly to [etherscan](https://etherscan.io)
```
qrcode echo "https://etherscan.io/pushTx?hex=0x$(ethsign send --to 0xffffffffffffffffffffffffffffffffffffffff --value 0.25 --key keyfile.key --nonce 42 -gasPrice 2 -gasLimit 21000)" > transaction.png
```


#### TODO

 - [x] Support non-ABI calls
 - [ ] Support arrays in non-ABI calls
 - [ ] Support multi-dimensional arrays in non-ABI calls
 - [ ] Tests
