package coin

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

const (
	MaxSizeInfoCoin   = 255
	JsonMarshalFlag   = 34 //nolint:revive
	CoinVersion1      = 1
	CoinVersion2      = 2
	TxRandomGroupSize = 68
)

const (
	PedersenPrivateKeyIndex = operation.PedersenPrivateKeyIndex
	PedersenValueIndex      = operation.PedersenValueIndex
	PedersenSndIndex        = operation.PedersenSndIndex
	PedersenShardIDIndex    = operation.PedersenShardIDIndex
	PedersenRandomnessIndex = operation.PedersenRandomnessIndex
)

func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseScalarForSetBytes(coinBytes *[]byte, offset *int) (*operation.Scalar, error) {
	b := *coinBytes
	if *offset >= len(b) {
		return nil, fmt.Errorf("offset is larger than len(bytes), cannot parse scalar")
	}
	var sc *operation.Scalar
	lenField := b[*offset]
	*offset++
	if lenField != 0 {
		if *offset+int(lenField) > len(b) {
			return nil, fmt.Errorf("offset+curLen is larger than len(bytes), cannot parse scalar for set bytes")
		}
		data := b[*offset : *offset+int(lenField)]
		sc = new(operation.Scalar).FromBytesS(data)
		*offset += int(lenField)
	}
	return sc, nil
}

func parsePointForSetBytes(coinBytes *[]byte, offset *int) (*operation.Point, error) {
	b := *coinBytes
	if *offset >= len(b) {
		return nil, fmt.Errorf("offset is larger than len(bytes), cannot parse point")
	}
	var point *operation.Point
	var err error
	lenField := b[*offset]
	*offset++
	if lenField != 0 {
		if *offset+int(lenField) > len(b) {
			return nil, fmt.Errorf("offset+curLen is larger than len(bytes), cannot parse point for set bytes")
		}
		data := b[*offset : *offset+int(lenField)]
		point, err = new(operation.Point).FromBytesS(data)
		if err != nil {
			return nil, err
		}
		*offset += int(lenField)
	}
	return point, nil
}

func parseInfoForSetBytes(coinBytes *[]byte, offset *int) ([]byte, error) {
	b := *coinBytes
	if *offset >= len(b) {
		return []byte{}, fmt.Errorf("offset is larger than len(bytes), cannot parse info")
	}
	info := []byte{}
	lenField := b[*offset]
	*offset++
	if lenField != 0 {
		if *offset+int(lenField) > len(b) {
			return []byte{}, fmt.Errorf("offset+curLen is larger than len(bytes), cannot parse info for set bytes")
		}
		info = make([]byte, lenField)
		copy(info, b[*offset:*offset+int(lenField)])
		*offset += int(lenField)
	}
	return info, nil
}

func CreatePaymentInfosFromPlainCoinsAndAddress(c []PlainCoin, paymentAddress key.PaymentAddress, message []byte) []*key.PaymentInfo {
	sumAmount := uint64(0)
	for i := 0; i < len(c); i++ {
		sumAmount += c[i].GetValue()
	}
	paymentInfos := make([]*key.PaymentInfo, 1)
	paymentInfos[0] = key.InitPaymentInfo(paymentAddress, sumAmount, message)
	return paymentInfos
}

func NewCoinFromAmountAndReceiver(amount uint64, receiver key.PaymentAddress) (*CoinV2, error) {
	paymentInfo := key.InitPaymentInfo(receiver, amount, []byte{})
	return NewCoinFromPaymentInfo((&CoinParams{}).FromPaymentInfo(paymentInfo))
}

func NewCoinFromAmountAndTxRandomBytes(amount uint64, publicKey *operation.Point, txRandom *TxRandom, info []byte) *CoinV2 {
	c := new(CoinV2).Init()
	c.SetPublicKey(publicKey)
	c.SetAmount(new(operation.Scalar).FromUint64(amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetTxRandom(txRandom)
	c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), c.GetRandomness(), operation.PedersenValueIndex))
	c.SetSharedRandom(nil)
	c.SetInfo(info)
	return c
}

func NewCoinFromPaymentInfo(p *CoinParams) (*CoinV2, error) {
	receiverPublicKey, err := new(operation.Point).FromBytesS(p.PaymentAddress.Pk)
	if err != nil {
		errStr := fmt.Sprintf("Cannot parse outputCoinV2 from PaymentInfo when parseByte PublicKey, error %v ", err)
		return nil, fmt.Errorf(errStr)
	}
	receiverPublicKeyBytes := receiverPublicKey.ToBytesS()
	targetShardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])

	c := new(CoinV2).Init()
	// Amount, Randomness, SharedRandom are transparency until we call concealData
	c.SetAmount(new(operation.Scalar).FromUint64(p.Amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetSharedRandom(operation.RandomScalar())        // shared randomness for creating one-time-address
	c.SetSharedConcealRandom(operation.RandomScalar()) // shared randomness for concealing amount and blinding asset tag
	c.SetInfo(p.Message)
	c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), c.GetRandomness(), operation.PedersenValueIndex))

	// If this is going to burning address then dont need to create ota
	if common.IsPublicKeyBurningAddress(p.PaymentAddress.Pk) {
		publicKey, err := new(operation.Point).FromBytesS(p.PaymentAddress.Pk)
		if err != nil {
			panic("Something is wrong with info.paymentAddress.pk, burning address should be a valid point")
		}
		c.SetPublicKey(publicKey)
		return c, nil
	}

	// Increase index until have the right shardID
	index := uint32(0)
	publicOTA := p.PaymentAddress.GetOTAPublicKey()
	if publicOTA == nil {
		return nil, fmt.Errorf("public OTA from payment address is nil")
	}
	publicSpend := p.PaymentAddress.GetPublicSpend()
	rK := new(operation.Point).ScalarMult(publicOTA, c.GetSharedRandom())
	for {
		index++

		// Get publickey
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := new(operation.Point).ScalarMultBase(hash)
		publicKey := new(operation.Point).Add(HrKG, publicSpend)
		c.SetPublicKey(publicKey)

		senderShardID, recvShardID, coinPrivacyType, _ := DeriveShardInfoFromCoin(publicKey.ToBytesS())
		if recvShardID == int(targetShardID) && senderShardID == p.SenderShardID && coinPrivacyType == p.CoinPrivacyType {
			otaRandomPoint := new(operation.Point).ScalarMultBase(c.GetSharedRandom())
			concealRandomPoint := new(operation.Point).ScalarMultBase(c.GetSharedConcealRandom())
			c.SetTxRandomDetail(concealRandomPoint, otaRandomPoint, index)

			b := append(rK.ToBytesS(), common.Uint32ToBytes(index)...)
			b = append(b, []byte("otatag")...)
			hOTATag := operation.HashToScalar(b)
			b = hOTATag.ToBytesS()
			vt := uint8(b[operation.Ed25519KeySize - 1])
			c.otaTag = &vt
			break
		}
	}
	return c, nil
}

func NewOTAFromReceiver(receiver key.PaymentAddress) (*operation.Point, *TxRandom, error) {
	paymentInfo := key.InitPaymentInfo(receiver, 0, []byte{})
	coin, err := NewCoinFromPaymentInfo((&CoinParams{}).FromPaymentInfo(paymentInfo))
	if err != nil {
		return nil, nil, err
	}
	return coin.GetPublicKey(), coin.txRandom, nil
}

func ParseOTAInfoToString(pubKey *operation.Point, txRandom *TxRandom) (string, string) {
	return base58.Base58Check{}.Encode(pubKey.ToBytesS(), common.ZeroByte), base58.Base58Check{}.Encode(txRandom.Bytes(), common.ZeroByte)
}

func ParseOTAInfoFromString(pubKeyStr, txRandomStr string) (*operation.Point, *TxRandom, error) {
	publicKeyB, version, err := base58.Base58Check{}.Decode(pubKeyStr)
	if err != nil || version != common.ZeroByte {
		return nil, nil, fmt.Errorf("parseOTAInfoFromString Cannot decode base58check string")
	}
	pubKey, err := new(operation.Point).FromBytesS(publicKeyB)
	if err != nil {
		return nil, nil, fmt.Errorf("parseOTAInfoFromString Cannot set Point from bytes")
	}

	txRandomB, version, err := base58.Base58Check{}.Decode(txRandomStr)
	if err != nil || version != common.ZeroByte {
		return nil, nil, fmt.Errorf("parseOTAInfoFromString Cannot decode base58check string")
	}
	txRandom := new(TxRandom)
	if err := txRandom.SetBytes(txRandomB); err != nil {
		return nil, nil, fmt.Errorf("parseOTAInfoFromString Cannot set txRandom from bytes")
	}
	return pubKey, txRandom, nil
}

const (
	PrivacyTypeTransfer = iota
	PrivacyTypeMint
)

// DeriveShardInfoFromCoin returns the sender origin & receiver shard of a coin based on the
// PublicKey on that coin (encoded inside its last byte).
// Does not support MaxShardNumber > 8
func DeriveShardInfoFromCoin(coinPubKey []byte) (int, int, int, error) {
	numShards := common.MaxShardNumber
	n := int(coinPubKey[len(coinPubKey)-1]) % 128 // use 7 bits
	receiverShardID := n % numShards
	n /= numShards
	senderShardID := n % numShards
	coinPrivacyType := n / numShards

	if coinPrivacyType > PrivacyTypeMint {
		return -1, -1, -1, fmt.Errorf("coin %x has unsupported PrivacyType %d", coinPubKey, coinPrivacyType)
	}
	return senderShardID, receiverShardID, coinPrivacyType, nil
}

// CoinParams contains the necessary data to create a new coin
type CoinParams struct {
	key.PaymentInfo
	SenderShardID   int
	CoinPrivacyType int
}

// From initializes the CoinParam using input data (PaymentInfo must not be nil)
func (p *CoinParams) From(inf *key.PaymentInfo, sid, cptype int) *CoinParams {
	return &CoinParams{
		PaymentInfo:     *inf,
		SenderShardID:   sid % common.MaxShardNumber,
		CoinPrivacyType: cptype % (PrivacyTypeMint + 1),
	}
}

// FromPaymentInfo initializes the CoinParam using a PaymentInfo (must not be nil);
// others are set to default
func (p *CoinParams) FromPaymentInfo(inf *key.PaymentInfo) *CoinParams {
	receiverPublicKeyBytes := inf.PaymentAddress.Pk
	shardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])
	return &CoinParams{
		PaymentInfo:     *inf,
		SenderShardID:   int(shardID),
		CoinPrivacyType: PrivacyTypeTransfer,
	}
}
