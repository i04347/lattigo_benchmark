// Package main implements an example of homomorphic LUT (Lookup Table) evaluation of the sign function using blind rotations implemented with the `rgsw` and `rgsw/lut` packages.
// These packages can be used to implement all the functionalities of the TFHE scheme.
package main

import (
	"fmt"
	"time"

	"github.com/tuneinsight/lattigo/v4/rgsw/lut"
	"github.com/tuneinsight/lattigo/v4/ring"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

// Function to evaluate
func sign(x float64) float64 {
	if x >= 0 {
		return 1
	}

	return -1
}

func main() {
	// RLWE parameters of the LUT
	// N=1024, Q=0x7fff801 -> ~2^128 ROP-security
	paramsLUT, err := rlwe.NewParametersFromLiteral(rlwe.ParametersLiteral{
		LogN:    10,
		Q:       []uint64{0x7fff801},
		NTTFlag: true,
	})

	if err != nil {
		panic(err)
	}

	// RLWE parameters of the samples
	// N=512, Q=0x3001 -> ~2^128 ROP-security
	paramsLWE, err := rlwe.NewParametersFromLiteral(rlwe.ParametersLiteral{
		LogN:    9,
		Q:       []uint64{0x3001},
		NTTFlag: true,
	})

	if err != nil {
		panic(err)
	}

	// Set the parameters for the blind rotation keys
	evkParams := rlwe.EvaluationKeyParameters{BaseTwoDecomposition: utils.Pointy(7)}

	// Scale of the RLWE samples
	scaleLWE := float64(paramsLWE.Q()[0]) / 4.0

	// Scale of the test poly
	scaleLUT := float64(paramsLUT.Q()[0]) / 4.0

	// Number of values samples stored in the RLWE sample
	slots := 32

	// Test poly
	LUTPoly := lut.InitLUT(sign, rlwe.NewScale(scaleLUT), paramsLUT.RingQ(), -1, 1)

	// Index map of which test poly to evaluate on which slot
	lutPolyMap := make(map[int]*ring.Poly)
	for i := 0; i < slots; i++ {
		lutPolyMap[i] = &LUTPoly
	}

	// RLWE secret for the samples
	skLWE := rlwe.NewKeyGenerator(paramsLWE).GenSecretKeyNew()

	// RLWE encryptor for the samples
	encryptorLWE, err := rlwe.NewEncryptor(paramsLWE, skLWE)
	if err != nil {
		panic(err)
	}

	// Values to encrypt in the RLWE sample
	values := make([]float64, slots)
	for i := 0; i < slots; i++ {
		values[i] = (-1.0 + float64(2*i)/float64(slots))
	}

	// Encode multiples values in a single RLWE
	ptLWE := rlwe.NewPlaintext(paramsLWE, paramsLWE.MaxLevel())
	for i := range values {
		if values[i] < 0 {
			ptLWE.Value.Coeffs[0][i] = paramsLWE.Q()[0] - uint64(-values[i]*scaleLWE)
		} else {
			ptLWE.Value.Coeffs[0][i] = uint64(values[i] * scaleLWE)
		}
	}

	paramsLWE.RingQ().NTT(ptLWE.Value, ptLWE.Value)

	// Encrypt the multiples values in a single RLWE
	ctLWE := rlwe.NewCiphertext(paramsLWE, 1, paramsLWE.MaxLevel())
	if err = encryptorLWE.Encrypt(ptLWE, ctLWE); err != nil {
		panic(err)
	}

	// Evaluator for the LUT evaluation
	eval := lut.NewEvaluator(paramsLUT, paramsLWE)

	// Secret of the RGSW ciphertexts encrypting the bits of skLWE
	skLUT := rlwe.NewKeyGenerator(paramsLUT).GenSecretKeyNew()

	// Collection of RGSW ciphertexts encrypting the bits of skLWE under skLUT
	blindeRotateKey, err := lut.GenEvaluationKeyNew(paramsLUT, skLUT, paramsLWE, skLWE, evkParams)
	if err != nil {
		panic(err)
	}

	// Evaluation of LUT(ctLWE)
	// Returns one RLWE sample per slot in ctLWE

	now := time.Now()
	ctsLUT, err := eval.Evaluate(ctLWE, lutPolyMap, blindeRotateKey)
	if err != nil{
		panic(err)
	}
	fmt.Printf("Done: %s (avg/LUT %3.1f [ms])\n", time.Since(now), float64(time.Since(now).Milliseconds())/float64(slots))

	// Decrypts, decodes and compares
	q := paramsLUT.Q()[0]
	qHalf := q >> 1
	decryptorLUT, err := rlwe.NewDecryptor(paramsLUT, skLUT)
	if err != nil {
		panic(err)
	}
	ptLUT := rlwe.NewPlaintext(paramsLUT, paramsLUT.MaxLevel())
	for i := 0; i < slots; i++ {

		decryptorLUT.Decrypt(ctsLUT[i], ptLUT)

		if ptLUT.IsNTT {
			paramsLUT.RingQ().INTT(ptLUT.Value, ptLUT.Value)
		}

		c := ptLUT.Value.Coeffs[0][0]

		var a float64
		if c >= qHalf {
			a = -float64(q-c) / scaleLUT
		} else {
			a = float64(c) / scaleLUT
		}

		fmt.Printf("%7.4f - %7.4f\n", a, values[i])
	}
}
