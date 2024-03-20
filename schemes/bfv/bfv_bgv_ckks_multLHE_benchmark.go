package main

import (
"fmt"
"os"
"time"
"github.com/tuneinsight/lattigo/v4/bfv"
"github.com/tuneinsight/lattigo/v4/rlwe"
)

func main() {
obliviousRiding()
}

func obliviousRiding() {
//defer profile.Start(profile.ProfilePath(".")).Stop()
//nbDrivers := 2048 //max is N

// BFV parameters (128 bit security) with plaintext modulus 65929217
paramDef := bfv.PN13QP218
//var btp *bootstrapping.Bootstrapper
//paramDef := bfv.PN15QP827pq
//paramDef := bfv.EXTRAPN11
//paramDef.T = 0x3ee0001
log, err := os.Create("top_bfv1")
if err != nil {
fmt.Println(err.Error())
}
defer log.Close()

//pwd := exec.Command("mpstat", "-P ALL 1").String()
//fmt.Printf("execute %s\n", pwd)
//cmd2 := exec.Command("mpstat", "-P", "ALL", "1")
/*cmd2 := exec.Command("top", "-b", "-d", "0.001")
pwd := cmd2.String()
fmt.Printf("execute %s\n", pwd)
cmd2.Stdout = log
cmd2.Stderr = log
if err := cmd2.Start(); err != nil {
fmt.Println(err.Error())
}
*/


//paramSet := bootstrapping.DefaultParametersSparse[0]

//btpParams := paramSet.BootstrappingParams
//paramDef = paramSet.SchemeParams

params, err := bfv.NewParametersFromLiteral(paramDef)
if err != nil {
panic(err)
}

//fmt.Println(params.QCount())
//for i := 0; i < params.QCount(); i++ {
// fmt.Println(params.Q()[i])
//}
encoder := bfv.NewEncoder(params)
start_key := time.Now()
// Rider's keygen
kgen := bfv.NewKeyGenerator(params)

riderSk, riderPk := kgen.GenKeyPair()
//fmt.Println("secret key size is ", riderSk.MarshalBinarySize(), "bytes.")
//fmt.Println("public key size is ", riderPk.MarshalBinarySize(), "bytes.")
rlk := kgen.GenRelinearizationKey(riderSk, 1)
//fmt.Println("relinearizationkey size is ", rlk.MarshalBinarySize(), "bytes.")
//riderSk, _ := kgen.GenKeyPair()
//_, riderPk := kgen.GenKeyPair()
end_key := time.Now()
fmt.Println("key generation time is ", end_key.Sub(start_key))

//encryptorRiderSk := bfv.NewEncryptor(params, riderSk)

evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk})

fmt.Println("finish key generation.")

fmt.Println("============================================")
fmt.Println("Homomorphic computations on batched integers")
fmt.Println("============================================")
fmt.Println()
//fmt.Printf("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
// 1<<params.LogN(), params.T(), params.LogQ(), params.Sigma())
fmt.Printf("Parameters : N=%d, Q = %d bits, sigma = %f \n",
1<<params.LogN(), params.LogQ(), params.Sigma())
fmt.Println()

fmt.Println()

const vector_size = 48
var p []uint64
//var p []float64
for i := 0; i < vector_size; i++ {
p = append(p, uint64(1))
//p = append(p, float64(1))
}
Plaintext := bfv.NewPlaintext(params, params.MaxLevel())
fmt.Println("level is ", params.MaxLevel())
encoder.Encode(p, Plaintext)
//encoder.Encode(p, Plaintext, 13)

/*valuesWant := make([]complex128, params.Slots())
for i := range valuesWant {
valuesWant[i] = utils.RandComplex128(-1, 1)
}

Plaintext := encoder.EncodeNew(valuesWant, params.MaxLevel(), params.DefaultScale(), params.LogSlots())
*/
encrypt_start := time.Now()
encryptorRiderPk := bfv.NewEncryptor(params, riderPk)
Ciphertext1 := encryptorRiderPk.EncryptNew(Plaintext)
encrypt_end := time.Now()
fmt.Println("encryption time is ", encrypt_end.Sub(encrypt_start))
Ciphertext2 := encryptorRiderPk.EncryptNew(Plaintext)
//Ciphertext2 := encryptorRiderPk.EncryptNew(Plaintext)
//fmt.Println("ciphertext size is ", Ciphertext1.MarshalBinarySize(), "bytes.")
/*log, err := os.Create("top_bfv")
if err != nil {
fmt.Println(err.Error())
}
defer log.Close()
*/
//pwd := exec.Command("mpstat", "-P ALL 1").String()
//fmt.Printf("execute %s\n", pwd)
//cmd2 := exec.Command("mpstat", "-P", "ALL", "1")

/*cmd2 := exec.Command("top", "-b", "-d", "0.001")
pwd := cmd2.String()
fmt.Printf("execute %s\n", pwd)
cmd2.Stdout = log
cmd2.Stderr = log
if err := cmd2.Start(); err != nil {
fmt.Println(err.Error())
}
time.Sleep(1 * time.Second)
*/


loop :=2
//r.Neg(Ciphertext1, Ciphertext1)
//if err := pprof.StartCPUProfile(f); err != nil {
//    return err
//}

start := time.Now()
for i := 0; i < loop; i++ {
//evaluator.Add(Ciphertext1, Ciphertext2, Ciphertext1)
//newcipher :=
//evaluator.Mul(Ciphertext1, Ciphertext2, Ciphertext1)
Ciphertext1 = evaluator.MulNew(Ciphertext1, Ciphertext2)
evaluator.Relinearize(Ciphertext1, Ciphertext1)
//evaluator.DropLevel(Ciphertext1, 1)
//rlwe.SwitchCiphertextRingDegree(newcipher, newcipher2)
//evaluator.Rescale(Ciphertext1, Ciphertext1.Scale, Ciphertext1) for ckks 同じ暗号文を繰り返し処理するとき
//evaluator.Rescale(Ciphertext1, Ciphertext1) bgvの場合のみ　bfvはmulの中でされるので必要ない  
//fmt.Println("level is ", params.MaxLevel())
//for bfv, bgv 同じ暗号文を繰り返し処理するとき
//Ciphertext1.Resize(Ciphertext1.Degree(), Ciphertext1.Level())
//sk2 := Ciphertext3.kgen.GenSecretKey()
//newchipher.Resize(Ciphertext1.Degree(), Ciphertext1.Level())
/*if Ciphertext1.Degree() > 3 {
fmt.Println(("in"))
newcipher2 := encryptorRiderPk.EncryptNew(Plaintext)
rlwe.SwitchCiphertextRingDegree(newcipher, newcipher2)
Ciphertext1 = newcipher2
continue
//Ciphertext1.Resize(Ciphertext1.Degree(), Ciphertext1.Level()-1)
}
*/
//Ciphertext1 = newcipher
//evaluator.(Ciphertext1,Ciphertext1)

}
end := time.Now()

elapsed := end.Sub(start)
fmt.Println("mult time is ")
fmt.Println(elapsed)
/*evk := bootstrapping.GenEvaluationKeys(btpParams, params, riderSk)
if btp, err = bootstrapping.NewBootstrapper(params, btpParams, evk); err != nil {
panic(err)
}

loop_bootstrap := 1
for i := 0; i < loop_bootstrap; i++ {
Ciphertext1 = btp.Bootstrap(Ciphertext1)
}
*/
decrypt_start := time.Now()
decryptor := bfv.NewDecryptor(params, riderSk)
result := encoder.DecodeUintNew(decryptor.DecryptNew(Ciphertext1))
//result := encoder.DecodeSlots(decryptor.DecryptNew(Ciphertext1), 13)
decrypt_end := time.Now()
fmt.Println("decryption time is ", decrypt_end.Sub(decrypt_start))
//result := encoder.DecodeUintNew(decryptor.DecryptNew(newchipher))
fmt.Println("result is ")
for i := 0; i < vector_size; i++ {
 fmt.Printf("%d ", result[i])
}

fmt.Println()

}

