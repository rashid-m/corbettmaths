package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/blockchain/types/mocks"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"
)

func TestValidateCommitteeSig(t *testing.T) {
	committee1 := []string{
		"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",
	}
	committeeStr1, _ := incognitokey.CommitteeBase58KeyListToStruct(committee1)
	valData1 := "{\"ProducerBLSSig\":\"Ts3ZvKYJ6hsUK3le0OFsHrC5VpsfYpAxoj0EG3q/Y2JGUwVpsy2Jr9zcJ2F9P1UnU9DKtOJUOb3/q09OluPEjAE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[0,1,2,3,4,5,6,7,8,9,10],\"AggSig\":\"obXe46HiZG+84h6rUUtcZLOyviHGErIqrjOw7ykKB9E=\",\"BridgeSig\":[\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\",\"\"]}"
	hash1, _ := common.Hash{}.NewHashFromStr("1a984591a379dcdce0bb4e75bfb8080234fd74c319387da0058326a8a98d84af")
	block1 := &mocks.BlockInterface{}
	block1.On("GetValidationField").Return(valData1).Once()
	block1.On("Hash").Return(hash1).Once()
	type args struct {
		block     types.BlockInterface
		committee []incognitokey.CommitteePublicKey
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "testcase 1",
			args: args{
				block:     block1,
				committee: committeeStr1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCommitteeSig(tt.args.block, tt.args.committee); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommitteeSig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeAndValidateSig(t *testing.T) {
	validationData := "{\"ProducerBLSSig\":\"bpnA8vEbSSntTxFcSiD+RHSoGDsCX8qCXDAuc2l41t4ypfQ9oNypU8MdXF/DUILW0wfzI+6BPArEHnw7ZoGFKQE=\",\"ProducerBriSig\":null,\"ValidatiorsIdx\":[1,2,3],\"AggSig\":\"kjVndVesMYMUz+Us7z98B1RzKWh4qSgvpxavlZ1yrRg=\",\"BridgeSig\":[\"\",\"\",\"\"]}"
	committeeKey := "121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM"

	valData, err := consensustypes.DecodeValidationData(validationData)
	if err != nil {
		t.Fatal(err)
	}

	producerKey := incognitokey.CommitteePublicKey{}
	err = producerKey.FromBase58(committeeKey)
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := common.Hash{}.NewHashFromStr("030bcfa39a9da5dc525c2424c44bf7a58480c5bb9e82f50c81a0f9a5093fc69c")
	if err := consensustypes.ValidateSingleBriSig(hash, valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		t.Fatal(err)
	}
}
