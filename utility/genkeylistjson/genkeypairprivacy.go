//+build !test

package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
	"io/ioutil"
	"os"
)

type AccountPub struct {
	PaymentAddress     string
	CommitteePublicKey string
}

const MAX_SHARD = 8
const MIN_MEMBER = 22

func main() {
	mnemonicGen := wallet.MnemonicGenerator{}
	//Entropy, _ := mnemonicGen.NewEntropy(128)
	//Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	Mnemonic := ""
	fmt.Printf("Mnemonic: %s\n", Mnemonic)
	Seed := mnemonicGen.NewSeed(Mnemonic, "dmnkajdklawjdkjawk")

	key, _ := wallet.NewMasterKey(Seed)
	var i = 0
	var j = 0

	listAcc := make(map[int][]AccountPub)
	listPrivAcc := make(map[int][]string)
	beaconAcc := []AccountPub{}
	beaconPriv := []string{}

	for j = 0; j < MAX_SHARD; j++ {
		listAcc[j] = []AccountPub{}
	}

	for {
		child, _ := key.NewChildKey(uint32(i))
		i++
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		committeeKeyStr, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)
		committeeKey, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committeeKeyStr})
		shardID := int(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
		if shardID >= MAX_SHARD {
			continue
		}

		if len(listAcc[shardID]) < MIN_MEMBER {
			listAcc[shardID] = append(listAcc[shardID], AccountPub{paymentAddress, committeeKey[0]})
			listPrivAcc[shardID] = append(listPrivAcc[shardID], privAddr)
		}

		shouldBreak := true
		for i, _ := range listAcc {
			if len(listAcc[i]) < MIN_MEMBER {
				shouldBreak = false
			}
		}

		if shouldBreak {
			break
		}
	}

	for {
		child, _ := key.NewChildKey(uint32(i))
		i++
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		committeeKeyStr, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)
		committeeKey, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committeeKeyStr})
		if len(beaconAcc) < MIN_MEMBER {
			beaconAcc = append(beaconAcc, AccountPub{paymentAddress, committeeKey[0]})
			beaconPriv = append(beaconPriv, privAddr)
		} else {
			break
		}

	}

	data, _ := json.Marshal(struct {
		Shard  map[int][]AccountPub
		Beacon []AccountPub
	}{
		listAcc,
		beaconAcc,
	})

	os.Remove("keylist.json")
	ioutil.WriteFile("keylist.json", data, 0x766)

	data, _ = json.Marshal(struct {
		Shard  map[int][]string
		Beacon []string
	}{
		listPrivAcc,
		beaconPriv,
	})

	os.Remove("priv.json")
	ioutil.WriteFile("priv.json", data, 0x766)

}
