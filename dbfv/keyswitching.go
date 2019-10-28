package dbfv

import (
	"github.com/ldsec/lattigo/bfv"
	"github.com/ldsec/lattigo/ring"
)

// CKSProtocol is a structure storing the parameters for the collective key-switching protocol.
type CKSProtocol struct {
	bfvContext *bfv.BfvContext

	sigmaSmudging         float64
	gaussianSamplerSmudge *ring.KYSampler

	tmpNtt   *ring.Poly
	tmpDelta *ring.Poly
	hP       *ring.Poly

	baseconverter *ring.FastBasisExtender
}

type CKSShare *ring.Poly

// NewCKSProtocol creates a new CKSProtocol that will be used to operate a collective key-switching on a ciphertext encrypted under a collective public-key, whose
// secret-shares are distributed among j parties, re-encrypting the ciphertext under an other public-key, whose secret-shares are also known to the
// parties.
func NewCKSProtocol(bfvContext *bfv.BfvContext, sigmaSmudging float64) *CKSProtocol {

	cks := new(CKSProtocol)

	cks.bfvContext = bfvContext

	cks.gaussianSamplerSmudge = bfvContext.ContextKeys().NewKYSampler(sigmaSmudging, int(6*sigmaSmudging))

	cks.tmpNtt = cks.bfvContext.ContextKeys().NewPoly()
	cks.tmpDelta = cks.bfvContext.ContextQ().NewPoly()
	cks.hP = cks.bfvContext.ContextPKeys().NewPoly()

	cks.baseconverter = ring.NewFastBasisExtender(cks.bfvContext.ContextQ().Modulus, cks.bfvContext.KeySwitchPrimes())

	return cks
}

func (cks *CKSProtocol) AllocateShare() CKSShare {
	return cks.bfvContext.ContextQ().NewPoly()
}

// GenShare is the first and unique round of the CKSProtocol protocol. Each party holding a ciphertext ctx encrypted under a collective publick-key musth
// compute the following :
//
// [(skInput_i - skOutput_i) * ctx[0] + e_i]
//
// Each party then broadcast the result of this computation to the other j-1 parties.
func (cks *CKSProtocol) GenShare(skInput, skOutput *ring.Poly, ct *bfv.Ciphertext, shareOut CKSShare) {

	cks.bfvContext.ContextQ().Sub(skInput, skOutput, cks.tmpDelta)

	cks.GenShareDelta(cks.tmpDelta, ct, shareOut)
}

func (cks *CKSProtocol) GenShareDelta(skDelta *ring.Poly, ct *bfv.Ciphertext, shareOut CKSShare) {

	level := uint64(len(ct.Value()[1].Coeffs) - 1)

	contextQ := cks.bfvContext.ContextQ()
	contextP := cks.bfvContext.ContextPKeys()

	contextQ.NTT(ct.Value()[1], cks.tmpNtt)
	contextQ.MulCoeffsMontgomery(cks.tmpNtt, skDelta, shareOut)

	for _, pj := range cks.bfvContext.KeySwitchPrimes() {
		contextQ.MulScalar(shareOut, pj, shareOut)
	}

	contextQ.InvNTT(shareOut, shareOut)

	cks.gaussianSamplerSmudge.Sample(cks.tmpNtt)
	contextQ.Add(shareOut, cks.tmpNtt, shareOut)

	for x, i := 0, uint64(len(contextQ.Modulus)); i < uint64(len(cks.bfvContext.ContextKeys().Modulus)); x, i = x+1, i+1 {
		for j := uint64(0); j < contextQ.N; j++ {
			cks.hP.Coeffs[x][j] += cks.tmpNtt.Coeffs[i][j]
		}
	}

	cks.baseconverter.ModDownSplited(contextQ, contextP, cks.bfvContext.RescaleParamsKeys(), level, shareOut, cks.hP, shareOut, cks.tmpNtt)

	cks.tmpNtt.Zero()
	cks.hP.Zero()
}

// AggregateShares is the second part of the unique round of the CKSProtocol protocol. Uppon receiving the j-1 elements each party computes :
//
// [ctx[0] + sum((skInput_i - skOutput_i) * ctx[0] + e_i), ctx[1]]
func (cks *CKSProtocol) AggregateShares(share1, share2, shareOut CKSShare) {
	cks.bfvContext.ContextQ().Add(share1, share2, shareOut)
}

// KeySwitch performs the actual keyswitching operation on a ciphertext ct and put the result in ctOut
func (cks *CKSProtocol) KeySwitch(combined CKSShare, ct *bfv.Ciphertext, ctOut *bfv.Ciphertext) {
	cks.bfvContext.ContextQ().Add(ct.Value()[0], combined, ctOut.Value()[0])
	cks.bfvContext.ContextQ().Copy(ct.Value()[1], ctOut.Value()[1])
}
