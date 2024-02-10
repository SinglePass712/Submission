package psetggm

/*
#cgo amd64 CXXFLAGS: -msse2 -msse -march=native -maes -Ofast -std=c++11 -I/opt/homebrew/include
#cgo arm64 CXXFLAGS: -march=armv8-a+fp+simd+crypto+crc -Ofast -std=c++11 -I/opt/homebrew/include 
#cgo LDFLAGS: -static-libstdc++ -L/opt/homebrew/lib -lcrypto
#include "pset_ggm.h"
#include "xor.h"
#include "answer.h"
#include "permute.h"
*/
import "C"
import (
	"unsafe"
)

type GGMSetGeneratorC struct {
	workspace []byte
	cgen      *C.generator
}

func NewGGMSetGeneratorC(univSize, setSize int) *GGMSetGeneratorC {
	size := C.workspace_size(C.uint(univSize), C.uint(setSize))
	gen := GGMSetGeneratorC{
		workspace: make([]byte, size),
	}
	gen.cgen = C.pset_ggm_init(C.uint(univSize), C.uint(setSize),
		(*C.uchar)(&gen.workspace[0]))
	return &gen
}

func (gen *GGMSetGeneratorC) Eval(seed []byte, elems []int) {
	C.pset_ggm_eval(gen.cgen, (*C.uchar)(&seed[0]), (*C.ulonglong)(unsafe.Pointer(&elems[0])))
}

func (gen *GGMSetGeneratorC) Punc(seed []byte, pos int) []byte {
	pset := make([]byte, C.pset_buffer_size(gen.cgen))
	C.pset_ggm_punc(gen.cgen, (*C.uchar)(&seed[0]), C.uint(pos), (*C.uchar)(&pset[0]))
	return pset
}

func (gen *GGMSetGeneratorC) EvalPunctured(pset []byte, hole int, elems []int) {
	C.pset_ggm_eval_punc(gen.cgen, (*C.uchar)(&pset[0]), C.uint(hole), (*C.ulonglong)(unsafe.Pointer(&elems[0])))
}

func XorBlocks(db []byte, offsets []int, out []byte) {
	C.xor_rows((*C.uchar)(&db[0]), C.uint(len(db)), (*C.ulonglong)(unsafe.Pointer(&offsets[0])), C.uint(len(offsets)), C.uint(len(out)), (*C.uchar)(&out[0]))
}

func XorBlocksTogether(db []byte, out []byte, elemSize int, numElems int) {
	C.xor_all_rows((*C.uchar)(&db[0]), C.uint(numElems), C.uint(elemSize),(*C.uchar)(&out[0]))

	//void xor_all_rows(const uint8_t* db, unsigned int db_len, unsigned int num_elems,unsigned int elem_size, uint8_t* out)
}

func XorHashesByBitVector(db []byte, indexing []byte, out []byte) {
	C.xor_hashes_by_bit_vector((*C.uchar)(&db[0]), C.uint(len(db)),
		(*C.uchar)(&indexing[0]), (*C.uchar)(&out[0]))
}

func (gen *GGMSetGeneratorC) Distinct(elems []int) bool {
	return (C.distinct(gen.cgen, (*C.ulonglong)(unsafe.Pointer(&elems[0])), C.uint(len(elems))) != 0)
}

func FastAnswer(pset []byte, hole, univSize, setSize, shift int, db []byte, rowLen int, out []byte) {
	C.answer((*C.uchar)(&pset[0]), C.uint(hole), C.uint(univSize), C.uint(setSize), C.uint(shift),
		(*C.uchar)(&db[0]), C.uint(len(db)), C.uint(rowLen), C.uint(len(out)), (*C.uchar)(&out[0]))
}

func SinglePassAnswer(db []byte, dbNumElems int,setNumElems int,dbElemSize int, 
	parities []byte, permSeed int, permutations []uint32, inverse_permutations []uint32) {
	C.answer_single_pass((*C.uchar)(&db[0]), C.uint(dbNumElems), C.uint(setNumElems), C.uint(dbElemSize),
		(*C.uchar)(&parities[0]), C.uint(permSeed), (*C.uint)(&permutations[0]), (*C.uint)(&inverse_permutations[0]))

}

func FastXorInto(out []byte, in []byte, elemSize int) {
	C.xor_into((*C.uchar)(&out[0]), (*C.uchar)(&in[0]), C.uint(elemSize))
}

func CopyIn(out []byte,db []byte, index int, elemSize int) {
	C.xor_into((*C.uchar)(&out[0]), (*C.uchar)(&db[index*elemSize]), C.uint(elemSize))	
	//C.xor_into((*C.uchar)(&out[0]), (*C.uchar)(&db[0]), C.uint(elemSize))	
}

func SinglePermutation(randomness uint, permArr []uint32, invPermArr []uint32, permSize uint32) {

	C.permute(C.uint(randomness),C.uint(permSize),(*C.uint)(&permArr[0]))
	C.invert_permutation((*C.uint)(&permArr[0]),C.uint(permSize),(*C.uint)(&invPermArr[0]))
}
