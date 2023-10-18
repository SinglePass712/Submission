package pir

import (
	"fmt"

	"github.com/lukechampine/fastxor"
)

type StaticDB struct {
	NumRows int
	RowLen  int
	Version int
	FlatDb  []byte
}

func (db *StaticDB) Slice(start, end int) []byte {
	return db.FlatDb[start*db.RowLen : end*db.RowLen]
}

func (db *StaticDB) Row(i int) Row {
	if i >= db.NumRows {
		return nil
	}
	return Row(db.Slice(i, i+1))
}

func StaticDBFromRows(data []Row) *StaticDB {
	if len(data) < 1 {
		return &StaticDB{0, 0,0,nil}
	}

	rowLen := len(data[0])
	flatDb := make([]byte, rowLen*len(data))

	for i, v := range data {
		if len(v) != rowLen {
			fmt.Printf("Got row[%v] %v %v\n", i, len(v), rowLen)
			panic("Database rows must all be of the same length")
		}

		copy(flatDb[i*rowLen:], v[:])
	}
	return &StaticDB{len(data), rowLen, 0,flatDb}
}

func (db StaticDB) Hint(req HintReq, resp *HintResp) (err error) {
	*resp, err = req.Process(db)
	return err
}
func (db StaticDB) Answer(q QueryReq, resp *interface{}) (err error) {
	*resp, err = q.Process(db)
	return err
}

/*-------------------------for backward compat-------------------------*/
func (db StaticDB) UpdateServer(updateIdxs []int, newRows []Row) error {
	return nil
}
func (db StaticDB) UpdateClient(u UpdateReq, resp *interface{}) (err error) {
	return nil
}
/*---------------------------------------------------------------------*/


type DBParams struct {
	NRows  int
	RowLen int
}

func (p *DBParams) NumRows() int {
	return p.NRows
}

func (db StaticDB) Params() *DBParams {
	return &DBParams{db.NumRows, db.RowLen}
}

func xorInto(a []byte, b []byte) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	fastxor.Bytes(a, a, b)

	// for i := 0; i < len(a); i++ {
	// 	a[i] = a[i] ^ b[i]
	// }
}


//new function for use with dynamicdb, simply changes certain row

func (db *StaticDB) UpdateRow(index int, updatedRow Row) {
	if index >= db.NumRows {
		//faulty update
		return
	}
	indFlatDb := index*db.RowLen
	for i := 0; i < db.RowLen; i++ {
		db.FlatDb[indFlatDb+i] = updatedRow[i]
	}

	db.Version +=1
}

func (db *StaticDB) ExtendDb(extendVal int) {
	currSize := len(db.FlatDb)
	for i:= currSize; i < currSize + (extendVal*db.RowLen); i++ {
		db.FlatDb = append(db.FlatDb, 0)
	}
	db.NumRows += extendVal
}
