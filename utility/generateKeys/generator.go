package main

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/utility/generateKeys/generator"
	"io/ioutil"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type KeyPairs struct {
	KeyPair []KeyPair
}

func main() {
	keys, _ := generator.GenerateAddressByShard(0)
	file, _ := json.MarshalIndent(keys, "", " ")
	_ = ioutil.WriteFile("private-keys-shard-0.json", file, 0644)
}
