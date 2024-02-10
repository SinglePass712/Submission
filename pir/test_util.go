package pir

import (
	"math/rand"
	"math"
)

var masterKey PRGKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 'A', 'B', 'C', 'D', 'E', 'F'}
var randReader *rand.Rand = rand.New(NewBufPRG(NewPRG(&masterKey)))

func RandSource() *rand.Rand {
	//return rand.New(rand.NewSource(17))
	return randReader
}

func MakeRows(src *rand.Rand, nRows, rowLen int) []Row {
	db := make([]Row, nRows)
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
		db[i][0] = byte(i % 256)
		db[i][1] = 'A' + byte(i%256)
	}
	return db
}

func MakeRandInts(src *rand.Rand, nInts, maxNum int) []int {
	randInts := make([]int, nInts)
	for i:= range randInts {
		randInts[i] = src.Intn(maxNum)
	}
	return randInts
}

func MakeDB(nRows int, rowLen int) StaticDB {
	return *StaticDBFromRows(MakeRows(RandSource(), nRows, rowLen))
}

func MakeDynamicDB(nRows int, rowLen int) DynamicDB {
	return *DynamicDBFromRows(MakeRows(RandSource(), nRows, rowLen), nRows, rowLen)
}

func MakeUpdateRows(nUpdates int, dbSize, rowLen int) ([]int, []Row) {
	updateInts := MakeRandInts(RandSource(), nUpdates, dbSize)
	updateRows := MakeRows(RandSource(), nUpdates, rowLen)
	
	return updateInts, updateRows

}


func MakeKeys(src *rand.Rand, nRows int) []uint32 {
	keys := make([]uint32, nRows)
	for i := range keys {
		keys[i] = uint32(src.Int31())
	}
	return keys
}

func MakeKeysRows(numRows, rowLen int) ([]uint32, []Row) {
	return MakeKeys(RandSource(), numRows), MakeRows(RandSource(), numRows, rowLen)
}

func GetSetSizeByCliSizeStaticSinglePass(numRows,rowLen,cliSize int) (setSize int) {
	prevI := 0
	for i := 1; i < numRows; i++ {
		if numRows % i == 0 {
			
			currCliSize := int(int(numRows/i) * rowLen) + int(float64(numRows)* (math.Log2(float64(numRows/i))/8))
			
			if (currCliSize < cliSize) {
				
				
					return i // cli storage always lower than checklist, can also return prevI to get slightly larger but still around same
				//}
			}
			prevI = i
		}
	}
	return prevI
}

func GetSetSizeByCliSizeDynamicSinglePass(numRows,rowLen,cliSize int) (setSize int) {
	prevI := 0
	for i := 1; i < numRows; i++ {
		if numRows % i == 0 {
			currCliSize := int(int(numRows/i) * rowLen) + int(float64(numRows)* (math.Log2(float64(numRows/i))/8)*2)
			
			if (currCliSize < cliSize) {
			//fmt.Printf("curr cli size: %d \n",currCliSize)	
					return i // cli storage always lower than checklist, can also return prevI to get slightly larger but still around same
				
				
			}
			prevI = i
		}
	}
	return prevI
}

func getChecklistBandwidth(numRows, rowLen int) (offlineBandwidth, onlineBandwidth int) {
	return int(math.Sqrt(float64(numRows)))*rowLen*int(math.Log(2)*128), 2*(rowLen + int(8 * math.Log2(float64(numRows)))) //128 bits = 16 bytes, 8 = 16/2 bc we actually only need log(N)/2
}
func getSinglePassBandwidth(numRows, rowLen,setSize int) (offlineBandwidth, onlineBandwidth int) {
	realSetSize := setSize
	if setSize == 0 {
		realSetSize = int(math.Sqrt(float64(numRows)))
	}

	numHints := int(numRows/realSetSize)
	return numHints*rowLen, 2*(rowLen*realSetSize + realSetSize*int(math.Log2(float64(numRows/realSetSize))/8)) //divided by two divided by bits/bytes
}
