package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"

	"math/big"

	"github.com/ninjadotorg/constant/privacy"
)

type MultiRangeProof struct {
	Counter byte
	Comms   []*privacy.EllipticPoint
	A       *privacy.EllipticPoint
	S       *privacy.EllipticPoint
	T1      *privacy.EllipticPoint
	T2      *privacy.EllipticPoint
	Tau     *big.Int
	Th      *big.Int
	Mu      *big.Int
	IPP     InnerProdArg
	maxExp  byte
	// challenges
	Cy *big.Int
	Cz *big.Int
	Cx *big.Int
}

type MultiRangeWitness struct {
	Comms  []*privacy.EllipticPoint
	Values []*big.Int
	Rands  []*big.Int
	maxExp byte
}

func (pro *MultiRangeProof) Init() *MultiRangeProof {
	pro.A = new(privacy.EllipticPoint).Zero()
	pro.S = new(privacy.EllipticPoint).Zero()
	pro.T1 = new(privacy.EllipticPoint).Zero()
	pro.T2 = new(privacy.EllipticPoint).Zero()
	pro.Tau = new(big.Int)
	pro.Th = new(big.Int)
	pro.Mu = new(big.Int)
	pro.Cx = new(big.Int)
	pro.Cy = new(big.Int)
	pro.Cz = new(big.Int)
	pro.IPP.A = new(big.Int)
	pro.IPP.B = new(big.Int)
	return pro
}

func (pro *MultiRangeProof) IsNil() bool {
	if pro.A == nil {
		return true
	}
	if pro.S == nil {
		return true
	}
	if pro.T1 == nil {
		return true
	}
	if pro.T2 == nil {
		return true
	}
	if pro.Tau == nil {
		return true
	}
	if pro.Th == nil {
		return true
	}
	if pro.Mu == nil {
		return true
	}
	if pro.Cx == nil {
		return true
	}
	if pro.Cy == nil {
		return true
	}
	if pro.Cz == nil {
		return true
	}
	if pro.IPP.A == nil {
		return true
	}
	if pro.IPP.B == nil {
		return true
	}
	return false
}

func (pro MultiRangeProof) Bytes() []byte {
	var res []byte

	if pro.IsNil() == true {
		return []byte{}
	}

	res = append(res, pro.Counter)
	res = append(res, pro.maxExp)

	for i := 0; i < int(pro.Counter); i++ {
		res = append(res, pro.Comms[i].Compress()...)
	}

	res = append(res, pro.A.Compress()...)
	res = append(res, pro.S.Compress()...)
	res = append(res, pro.T1.Compress()...)
	res = append(res, pro.T2.Compress()...)

	res = append(res, privacy.AddPaddingBigInt(pro.Tau, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(pro.Th, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(pro.Mu, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(pro.Cx, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(pro.Cy, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(pro.Cz, privacy.BigIntSize)...)
	res = append(res, pro.IPP.Bytes()...)
	return res

}

func (pro *MultiRangeProof) SetBytes(proofbytes []byte) error {
	if pro.IsNil() {
		pro = pro.Init()
	}

	if len(proofbytes) == 0 {
		return nil
	}

	pro.Counter = proofbytes[0]
	pro.maxExp = proofbytes[1]
	offset := 2

	pro.Comms = make([]*privacy.EllipticPoint, pro.Counter)
	for i := 0; i < int(pro.Counter); i++ {
		pro.Comms[i] = new(privacy.EllipticPoint)
		err := pro.Comms[i].Decompress(proofbytes[offset: offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset += privacy.CompressedPointSize
	}

	pro.A = new(privacy.EllipticPoint)
	err := pro.A.Decompress(proofbytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.S = new(privacy.EllipticPoint)
	err = pro.S.Decompress(proofbytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.T1 = new(privacy.EllipticPoint)
	err = pro.T1.Decompress(proofbytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.T2 = new(privacy.EllipticPoint)
	err = pro.T2.Decompress(proofbytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.Tau = new(big.Int)
	pro.Tau.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	pro.Th = new(big.Int)
	pro.Th.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	pro.Mu = new(big.Int)
	pro.Mu.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	pro.Cx = new(big.Int)
	pro.Cx.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	pro.Cy = new(big.Int)
	pro.Cy.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	pro.Cz = new(big.Int)
	pro.Cz.SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize
	end := len(proofbytes)
	pro.IPP.SetBytes(proofbytes[offset:end])
	return nil
}

func initCommonParams(l int, maxExp byte) {
	VecLength = int(maxExp) * pad(l)
	RangeProofParams = newECPrimeGroupKey(VecLength)
}

func (wit *MultiRangeWitness) Set(v []*big.Int, maxExp byte) {
	l := pad(len(v) + 1)
	wit.Values = make([]*big.Int, l)
	for i := 0; i < l; i++ {
		wit.Values[i] = new(big.Int)
		wit.Values[i].SetInt64(0)
	}
	total := new(big.Int).SetUint64(0)
	for i := 0; i < len(v); i++ {
		wit.Values[i] = new(big.Int)
		*wit.Values[i] = *v[i]
		total.Add(total, v[i])
	}
	*wit.Values[l-1] = *total
	wit.maxExp = maxExp
}

// Calculates (aL - z*1^n) + sL*x
func calculateLMRP(aL, sL []*big.Int, z, x *big.Int) []*big.Int {
	result := make([]*big.Int, len(aL))
	tmp1 := VectorAddScalar(aL, new(big.Int).Neg(z))
	tmp2 := ScalarVectorMul(sL, x)
	result = VectorAdd(tmp1, tmp2)
	return result
	//return nil
}

func calculateRMRP(aR, sR, y, zTimesTwo []*big.Int, z, x *big.Int) []*big.Int {
	if len(aR) != len(sR) || len(aR) != len(y) || len(y) != len(zTimesTwo) {
		return nil
	}
	result := make([]*big.Int, len(aR))
	tmp11 := VectorAddScalar(aR, z)
	tmp12 := ScalarVectorMul(sR, x)
	tmp13 := VectorAdd(tmp11, tmp12)
	tmp1 := VectorHadamard(y, tmp13)
	result = VectorAdd(tmp1, zTimesTwo)
	return result
}

/*
DeltaMRP is a helper function that is used in the multi range proof

\delta(y, z) = (z-z^2)<1^n, y^n> - \sum_j z^3+j<1^n, 2^n>
*/
func deltaMRP(y []*big.Int, z *big.Int, m int) *big.Int {
	result := big.NewInt(0)
	// (z-z^2)<1^n, y^n>
	z2 := new(big.Int).Mod(new(big.Int).Mul(z, z), privacy.Curve.Params().N)
	t1 := new(big.Int).Mod(new(big.Int).Sub(z, z2), privacy.Curve.Params().N)
	t2 := new(big.Int).Mod(new(big.Int).Mul(t1, VectorSum(y)), privacy.Curve.Params().N)

	// \sum_j z^3+j<1^n, 2^n>
	// <1^n, 2^n> = 2^n - 1
	po2sum := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(RangeProofParams.V/m)), privacy.Curve.Params().N), big.NewInt(1))
	t3 := big.NewInt(0)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(z, big.NewInt(3+int64(j)), privacy.Curve.Params().N)
		tmp1 := new(big.Int).Mod(new(big.Int).Mul(zp, po2sum), privacy.Curve.Params().N)
		t3 = new(big.Int).Mod(new(big.Int).Add(t3, tmp1), privacy.Curve.Params().N)
	}
	result = new(big.Int).Mod(new(big.Int).Sub(t2, t3), privacy.Curve.Params().N)
	return result
}

/*
MultiRangeProof Prove
Takes in a list of values and provides an aggregate
range proof for all the values.

changes:
 all values are concatenated
 r(x) is computed differently
 tau_x calculation is different
 delta calculation is different

{(g, h \in G, \textbf{V} \in G^m ; \textbf{v, \gamma} \in Z_p^m) :
	V_j = h^{\gamma_j}g^{v_j} \wedge v_j \in [0, 2^n - 1] \forall j \in [1, m]}
*/
func (wit *MultiRangeWitness) Prove() (*MultiRangeProof, error) {
	// RangeProofParams.V has the total number of values and bits we can support

	initCommonParams(len(wit.Values), wit.maxExp)
	MRProof := MultiRangeProof{}
	MRProof.maxExp = wit.maxExp
	m := len(wit.Values)
	MRProof.Counter = byte(m)
	bitsPerValue := RangeProofParams.V / m

	// we concatenate the binary representation of the values
	PowerOfTwos := PowerVector(bitsPerValue, big.NewInt(2))
	Comms := make([]*privacy.EllipticPoint, m)
	gammas := make([]*big.Int, m)
	wit.Rands = make([]*big.Int, m)
	aLConcat := make([]*big.Int, RangeProofParams.V)
	aRConcat := make([]*big.Int, RangeProofParams.V)
	sumRand := new(big.Int)
	sumRand.SetUint64(0)
	for j := range wit.Values {
		v := wit.Values[j]
		if v.Cmp(big.NewInt(0)) == -1 {
			return nil, errors.New("Value is below range")
		}
		if v.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(wit.maxExp)), privacy.Curve.Params().N)) == 1 {
			return nil, errors.New("Value is above range")
		}
		gamma := new(big.Int).SetBytes(privacy.RandBytes(32))
		gamma.Mod(gamma, privacy.Curve.Params().N)
		Comms[j] = RangeProofParams.G.ScalarMult(v).Add(RangeProofParams.H.ScalarMult(gamma))
		gammas[j] = gamma
		wit.Rands[j] = gamma
		// break up v into its bitwise representation
		aL := reverse(StrToBigIntArray(PadLeft(fmt.Sprintf("%b", v), "0", bitsPerValue)))
		aR := VectorAddScalar(aL, big.NewInt(-1))
		for i := range aR {
			aLConcat[bitsPerValue*j+i] = aL[i]
			aRConcat[bitsPerValue*j+i] = aR[i]
		}
	}
	MRProof.Comms = Comms
	wit.Comms = Comms
	alpha := new(big.Int).SetBytes(privacy.RandBytes(32))
	alpha.Mod(alpha, privacy.Curve.Params().N)

	A := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, aLConcat, aRConcat).Add(RangeProofParams.H.ScalarMult((alpha)))
	if (A == nil) {
		return nil, errors.New("Creating multi-range proof failed")
	} else {
		MRProof.A = A
	}

	sL := RandVector(RangeProofParams.V)
	sR := RandVector(RangeProofParams.V)

	rho := new(big.Int).SetBytes(privacy.RandBytes(32))
	rho.Mod(alpha, privacy.Curve.Params().N)

	S := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, sL, sR).Add(RangeProofParams.H.ScalarMult(rho))
	if (S == nil) {
		return nil, errors.New("Creating multi-range proof failed")
	} else {
		MRProof.S = S
	}
	chal1s256 := common.HashB([]byte(A.X.String() + A.Y.String()))
	cy := new(big.Int).SetBytes(chal1s256[:])
	MRProof.Cy = cy

	chal2s256 := common.HashB([]byte(S.X.String() + S.Y.String()))
	cz := new(big.Int).SetBytes(chal2s256[:])
	MRProof.Cz = cz

	zPowersTimesTwoVec := make([]*big.Int, RangeProofParams.V)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		for i := 0; i < bitsPerValue; i++ {
			zPowersTimesTwoVec[j*bitsPerValue+i] = new(big.Int).Mod(new(big.Int).Mul(PowerOfTwos[i], zp), privacy.Curve.Params().N)
		}
	}
	// need to generate l(X), r(X), and t(X)=<l(X),r(X)>
	PowerOfCY := PowerVector(RangeProofParams.V, cy)
	l0 := VectorAddScalar(aLConcat, new(big.Int).Neg(cz))
	l1 := sL
	r0 := VectorAdd(
		VectorHadamard(
			PowerOfCY,
			VectorAddScalar(aRConcat, cz)),
		zPowersTimesTwoVec)
	if (r0 == nil) {
		return nil, errors.New("Creating multi-range proof failed")
	}
	r1 := VectorHadamard(sR, PowerOfCY)

	//calculate t0
	vz2 := big.NewInt(0)
	z2 := new(big.Int).Mod(new(big.Int).Mul(cz, cz), privacy.Curve.Params().N)
	PowerOfCZ := PowerVector(m, cz)
	for j := 0; j < m; j++ {
		vz2 = new(big.Int).Add(vz2,
			new(big.Int).Mul(
				PowerOfCZ[j],
				new(big.Int).Mul(wit.Values[j], z2)))
		vz2 = new(big.Int).Mod(vz2, privacy.Curve.Params().N)
	}

	t0 := new(big.Int).Mod(new(big.Int).Add(vz2, deltaMRP(PowerOfCY, cz, m)), privacy.Curve.Params().N)

	t1 := new(big.Int).Mod(new(big.Int).Add(InnerProduct(l1, r0), InnerProduct(l0, r1)), privacy.Curve.Params().N)
	t2 := InnerProduct(l1, r1)
	if (t2 == nil) {
		return nil, errors.New("Creating multi-range proof failed")
	}
	// given the t_i values, we can generate commitments to them
	tau1 := new(big.Int).SetBytes(privacy.RandBytes(32))
	tau1.Mod(tau1, privacy.Curve.Params().N)

	tau2 := new(big.Int).SetBytes(privacy.RandBytes(32))
	tau2.Mod(tau2, privacy.Curve.Params().N)

	T1 := RangeProofParams.G.ScalarMult(t1).Add(RangeProofParams.H.ScalarMult(tau1)) //commitment to t1
	T2 := RangeProofParams.G.ScalarMult(t2).Add(RangeProofParams.H.ScalarMult(tau2)) //commitment to t2

	MRProof.T1 = T1
	MRProof.T2 = T2

	chal3s256 := common.HashB([]byte(T1.X.String() + T1.Y.String() + T2.X.String() + T2.Y.String()))
	cx := new(big.Int).SetBytes(chal3s256[:])

	MRProof.Cx = cx

	left := calculateLMRP(aLConcat, sL, cz, cx)
	right := calculateRMRP(aRConcat, sR, PowerOfCY, zPowersTimesTwoVec, cz, cx)

	thatPrime := new(big.Int).Mod( // t0 + t1*x + t2*x^2
		new(big.Int).Add(t0, new(big.Int).Add(new(big.Int).Mul(t1, cx), new(big.Int).Mul(new(big.Int).Mul(cx, cx), t2))), privacy.Curve.Params().N)

	that := InnerProduct(left, right) // NOTE: BP Java implementation calculates this from the t_i
	// thatPrime and that should be equal
	if thatPrime.Cmp(that) != 0 {
		return nil, errors.New("Creating multi-range proof failed")
	}
	MRProof.Th = that
	vecRandomnessTotal := big.NewInt(0)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		tmp1 := new(big.Int).Mul(gammas[j], zp)
		vecRandomnessTotal = new(big.Int).Mod(new(big.Int).Add(vecRandomnessTotal, tmp1), privacy.Curve.Params().N)
	}
	taux1 := new(big.Int).Mod(new(big.Int).Mul(tau2, new(big.Int).Mul(cx, cx)), privacy.Curve.Params().N)
	taux2 := new(big.Int).Mod(new(big.Int).Mul(tau1, cx), privacy.Curve.Params().N)
	taux := new(big.Int).Mod(new(big.Int).Add(taux1, new(big.Int).Add(taux2, vecRandomnessTotal)), privacy.Curve.Params().N)
	MRProof.Tau = taux
	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(rho, cx)), privacy.Curve.Params().N)
	MRProof.Mu = mu
	HPrime := make([]*privacy.EllipticPoint, len(RangeProofParams.BPH))
	for i := range HPrime {
		HPrime[i] = RangeProofParams.BPH[i].ScalarMult(new(big.Int).ModInverse(PowerOfCY[i], privacy.Curve.Params().N))
	}
	P := TwoVectorPCommitWithGens(RangeProofParams.BPG, HPrime, left, right)
	if (P == nil) {
		return nil, errors.New("Creating multi-range proof failed")
	}
	MRProof.IPP = InnerProductProve(left, right, that, P, RangeProofParams.U, RangeProofParams.BPG, HPrime)
	return &MRProof, nil
}

/*
MultiRangeProof Verify
Takes in a MultiRangeProof and verifies its correctness
*/
func (pro *MultiRangeProof) Verify() bool {
	m := len(pro.Comms)
	initCommonParams(m, pro.maxExp)
	if m == 0 {
		return false
	}
	bitsPerValue := RangeProofParams.V / m
	//changes:
	// check 1 changes since it includes all commitments
	// check 2 commitment generation is also different
	// verify the challenges
	chal1s256 := common.HashB([]byte(pro.A.X.String() + pro.A.Y.String()))
	cy := new(big.Int).SetBytes(chal1s256[:])
	if cy.Cmp(pro.Cy) != 0 {
		return false
	}
	chal2s256 := common.HashB([]byte(pro.S.X.String() + pro.S.Y.String()))
	cz := new(big.Int).SetBytes(chal2s256[:])
	if cz.Cmp(pro.Cz) != 0 {
		return false
	}
	chal3s256 := common.HashB([]byte(pro.T1.X.String() + pro.T1.Y.String() + pro.T2.X.String() + pro.T2.Y.String()))
	cx := new(big.Int).SetBytes(chal3s256[:])
	if cx.Cmp(pro.Cx) != 0 {
		return false
	}
	// given challenges are correct, very range proof
	PowersOfY := PowerVector(RangeProofParams.V, cy)
	// t_hat * G + tau * H
	lhs := RangeProofParams.G.ScalarMult(pro.Th).Add(RangeProofParams.H.ScalarMult(pro.Tau))
	// z^2 * \bold{z}^m \bold{V} + delta(y,z) * G + x * T1 + x^2 * T2
	CommPowers := RangeProofParams.Zero()
	PowersOfZ := PowerVector(m, cz)
	z2 := new(big.Int).Mod(new(big.Int).Mul(cz, cz), privacy.Curve.Params().N)

	for j := 0; j < m; j++ {
		CommPowers = CommPowers.Add(pro.Comms[j].ScalarMult(new(big.Int).Mul(z2, PowersOfZ[j])))
	}
	rhs := RangeProofParams.G.ScalarMult(deltaMRP(PowersOfY, cz, m)).Add(
		pro.T1.ScalarMult(cx)).Add(
		pro.T2.ScalarMult(new(big.Int).Mul(cx, cx))).Add(CommPowers)

	if !lhs.IsEqual(rhs) {
		return false
	}

	tmp1 := RangeProofParams.Zero()
	zneg := new(big.Int).Mod(new(big.Int).Neg(cz), privacy.Curve.Params().N)
	for i := range RangeProofParams.BPG {
		tmp1 = tmp1.Add(RangeProofParams.BPG[i].ScalarMult(zneg))
	}
	PowerOfTwos := PowerVector(bitsPerValue, big.NewInt(2))
	tmp2 := RangeProofParams.Zero()
	// generate h'
	HPrime := make([]*privacy.EllipticPoint, len(RangeProofParams.BPH))

	for i := range HPrime {
		mi := new(big.Int).ModInverse(PowersOfY[i], privacy.Curve.Params().N)
		HPrime[i] = RangeProofParams.BPH[i].ScalarMult(mi)
	}

	for j := 0; j < m; j++ {
		for i := 0; i < bitsPerValue; i++ {
			val1 := new(big.Int).Mul(cz, PowersOfY[j*bitsPerValue+i])
			zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
			val2 := new(big.Int).Mod(new(big.Int).Mul(zp, PowerOfTwos[i]), privacy.Curve.Params().N)
			tmp2 = tmp2.Add(HPrime[j*bitsPerValue+i].ScalarMult(new(big.Int).Add(val1, val2)))
		}
	}
	tmp, _ := RangeProofParams.H.ScalarMult(pro.Mu).Inverse()
	P := pro.A.Add(pro.S.ScalarMult(cx)).Add(tmp1).Add(tmp2).Add(tmp)
	if !InnerProductVerifyFast(pro.Th, P, RangeProofParams.U, RangeProofParams.BPG, HPrime, pro.IPP) {
		return false
	}
	return true
}
