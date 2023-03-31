package rlwe

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/google/go-cmp/cmp"
	"github.com/tuneinsight/lattigo/v4/utils/buffer"
)

// GaloisKey is a type of evaluation key used to evaluate automorphisms on ciphertext.
// An automorphism pi: X^{i} -> X^{i*GaloisElement} changes the key under which the
// ciphertext is encrypted from s to pi(s). Thus, the ciphertext must be re-encrypted
// from pi(s) to s to ensure correctness, which is done with the corresponding GaloisKey.
//
// Lattigo implements automorphismes differently than the usual way (which is to first
// apply the automorphism and then the evaluation key). Instead the order of operations
// is reversed, the GaloisKey for pi^{-1} is evaluated on the ciphertext, outputing a
// ciphertext encrypted under pi^{-1}(s), and then the automorphism pi is applied. This
// enables a more efficient evaluation, by only having to apply the automorphism on the
// final result (instead of having to apply it on the decomposed ciphertext).
type GaloisKey struct {
	GaloisElement uint64
	NthRoot       uint64
	EvaluationKey
}

// NewGaloisKey allocates a new GaloisKey with zero coefficients and GaloisElement set to zero.
func NewGaloisKey(params Parameters) *GaloisKey {
	return &GaloisKey{EvaluationKey: *NewEvaluationKey(params, params.MaxLevelQ(), params.MaxLevelP()), NthRoot: params.RingQ().NthRoot()}
}

// Equal returns true if the two objects are equal.
func (gk *GaloisKey) Equal(other *GaloisKey) bool {
	return gk.GaloisElement == other.GaloisElement && gk.NthRoot == other.NthRoot && cmp.Equal(gk.EvaluationKey, other.EvaluationKey)
}

// CopyNew creates a deep copy of the object and returns it
func (gk *GaloisKey) CopyNew() *GaloisKey {
	return &GaloisKey{
		GaloisElement: gk.GaloisElement,
		NthRoot:       gk.NthRoot,
		EvaluationKey: *gk.EvaluationKey.CopyNew(),
	}
}

// BinarySize returns the size in bytes that the object once marshalled into a binary form.
func (gk *GaloisKey) BinarySize() (dataLen int) {
	return gk.EvaluationKey.BinarySize() + 16
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (gk *GaloisKey) MarshalBinary() (data []byte, err error) {
	data = make([]byte, gk.BinarySize())
	_, err = gk.Read(data)
	return
}

// Read encodes the object into a binary form on a preallocated slice of bytes
// and returns the number of bytes written.
func (gk *GaloisKey) Read(data []byte) (ptr int, err error) {

	if len(data) < 16 {
		return ptr, fmt.Errorf("cannot write: len(data) < 16")
	}

	binary.LittleEndian.PutUint64(data[ptr:], gk.GaloisElement)
	ptr += 8

	binary.LittleEndian.PutUint64(data[ptr:], gk.NthRoot)
	ptr += 8

	var inc int
	if inc, err = gk.EvaluationKey.Read(data[ptr:]); err != nil {
		return
	}

	ptr += inc

	return
}

// WriteTo writes the object on an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Writer, which defines
// a subset of the method of the bufio.Writer.
// If w is not compliant to the buffer.Writer interface, it will be wrapped in
// a new bufio.Writer.
// For additional information, see lattigo/utils/buffer/writer.go.
func (gk *GaloisKey) WriteTo(w io.Writer) (n int64, err error) {
	switch w := w.(type) {
	case buffer.Writer:

		var inc int

		if inc, err = buffer.WriteUint64(w, gk.GaloisElement); err != nil {
			return n + int64(inc), err
		}

		n += int64(inc)

		if inc, err = buffer.WriteUint64(w, gk.NthRoot); err != nil {
			return n + int64(inc), err
		}

		n += int64(inc)

		var inc2 int64
		if inc2, err = gk.EvaluationKey.WriteTo(w); err != nil {
			return n + inc2, err
		}

		n += inc2

		return

	default:
		return gk.WriteTo(bufio.NewWriter(w))
	}
}

// UnmarshalBinary decodes a slice of bytes generated by MarshalBinary
// or Read on the object.
func (gk *GaloisKey) UnmarshalBinary(data []byte) (err error) {
	_, err = gk.Write(data)
	return
}

// ReadFrom reads on the object from an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Reader, which defines
// a subset of the method of the bufio.Reader.
// If r is not compliant to the buffer.Reader interface, it will be wrapped in
// a new bufio.Reader.
// For additional information, see lattigo/utils/buffer/reader.go.
func (gk *GaloisKey) ReadFrom(r io.Reader) (n int64, err error) {
	switch r := r.(type) {
	case buffer.Reader:

		var inc int

		if inc, err = buffer.ReadUint64(r, &gk.GaloisElement); err != nil {
			return n + int64(inc), err
		}

		n += int64(inc)

		if inc, err = buffer.ReadUint64(r, &gk.NthRoot); err != nil {
			return n + int64(inc), err
		}

		n += int64(inc)

		var inc2 int64
		if inc2, err = gk.EvaluationKey.ReadFrom(r); err != nil {
			return n + inc2, err
		}

		n += inc2

		return
	default:
		return gk.ReadFrom(bufio.NewReader(r))
	}
}

// Write decodes a slice of bytes generated by MarshalBinary or
// Read on the object and returns the number of bytes read.
func (gk *GaloisKey) Write(data []byte) (ptr int, err error) {

	if len(data) < 16 {
		return ptr, fmt.Errorf("cannot read: len(data) < 16")
	}

	gk.GaloisElement = binary.LittleEndian.Uint64(data[ptr:])
	ptr += 8

	gk.NthRoot = binary.LittleEndian.Uint64(data[ptr:])
	ptr += 8

	var inc int
	if inc, err = gk.EvaluationKey.Write(data[ptr:]); err != nil {
		return
	}

	ptr += inc

	return
}
