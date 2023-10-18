package pir

import (
	"fmt"

	// "github.com/lukechampine/fastxor"
	"math"
	"time"
)

type UpdateOp struct {
	Index    int
	OldData  Row     //need old data to 'xor out' of hint, alternatively if versioning it is not necessary to store directly
					 //client can pass out version and one can look at updates since such version

	NewData  Row     //since we are actually performing an update, we can just store a map and don't need to store new
					 //data (can populate at 'update' query time). only constant improvement so this seems cleaner. Again
					 //if versioning the database then this is unnecessary only need to keep track of diff indices
}

type DynamicDB struct {
	staticDb StaticDB
	ops []UpdateOp
	currVersion int
	opInitVersion int
	MaxOpSize int
	UnloadOpsRate int
	realSizeDB int
	extendSize int
}


func DynamicDBFromRows(rows []Row, numRows int, rowSize int) *DynamicDB {
	db := *StaticDBFromRows(rows)
	ops := make([]UpdateOp,0)
	maxOpSize := 10000//get setting from somewhere
	unloadOpsRate := 0//get setting from somewhere
	extendSize := int(numRows/int(math.Sqrt(float64(db.NumRows))))//get setting from somewhere (must equal numRows/setSize)
	return &DynamicDB{db,ops,0,0,maxOpSize,unloadOpsRate,numRows,extendSize}
}

func (ddb *DynamicDB) Row(i int) Row {
	return ddb.staticDb.Row(i)
}

//TODO: finish out static db ideas (can be deamortized)

//Have DynamicPIRreader class that inits with dynamicdb rather than static!
// (to benchmark against benchmark_initial and benchmark_trace for checklist)


func (ddb *DynamicDB) AppendRows(newRows []Row) error{

	ddb.checkUpdateState()


	numAdds := len(newRows)

	if ddb.realSizeDB + numAdds >= ddb.staticDb.NumRows {
		ddb.extendStaticDb()
	}


	rangeArr := makeRangeArray(ddb.realSizeDB, ddb.realSizeDB+numAdds)
	err := ddb.EditRows(rangeArr,newRows)
	if err != nil {
		return err
	}

	ddb.realSizeDB += numAdds
	return nil

}

func (ddb *DynamicDB) EditRows(updateIndices []int, updatedRows []Row) error {

	if (len(updateIndices) != len(updatedRows)) {

		//faulty update
		return fmt.Errorf("Input arrays are of different sizes")
	}

	ddb.checkUpdateState()

	for i := 0; i < len(updateIndices); i++ {
		editIndex := updateIndices[i]
		//append to ops ('diff' analog)
		oldVal := make(Row, len(updatedRows[0]))
		copy(oldVal, ddb.staticDb.Row(editIndex))
		//newVal := updatedRows[i]

		ddb.ops = append(ddb.ops, UpdateOp{editIndex,oldVal, updatedRows[i]})

		//update staticDb
		ddb.staticDb.UpdateRow(editIndex, updatedRows[i])

	}

	ddb.currVersion += len(updateIndices)
	return nil

}


func (ddb *DynamicDB) checkUpdateState() {
	if (len(ddb.ops) > ddb.MaxOpSize) {
		//remove oldest ops (by the unloadOpsRate number)
		//update opInitVersion

	}
}

func (ddb *DynamicDB) extendStaticDb() {
	ddb.staticDb.ExtendDb(ddb.extendSize)
}




//pir_reader/test funcs


func (ddb *DynamicDB) Hint(req HintReq, resp *HintResp) (err error) {
	*resp, err = req.Process(ddb.staticDb)
	return err
}
func (ddb *DynamicDB) Answer(q QueryReq, resp *interface{}) (err error) {
	*resp, err = q.Process(ddb.staticDb)
	return err
}
func (ddb *DynamicDB) UpdateServer(updateIdxs []int, newRows []Row) error {
	return ddb.EditRows(updateIdxs,newRows)
}

func (ddb *DynamicDB) UpdateClient(u UpdateReq, resp *interface{}) (err error) {
	start := time.Now()
	*resp, err = u.GetUpdates(*ddb)
	elapsed:= time.Since(start)
	fmt.Printf("update server time: %d ns \n", elapsed.Nanoseconds())
	return err
}



//util

func makeRangeArray(min, max int) []int {
    a := make([]int, max-min)
    for i := range a {
        a[i] = min + i
    }
    return a
}