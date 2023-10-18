package pir

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"
	"fmt"
	"time"
	// "math"
)

func TestSinglePassUpdatableFixedSize(t *testing.T) {
	if testing.Short() {
  		t.Skip("skipping testing in short mode")
	}
	checklistCliSizes := [][]int{{7014578, 14915058, 38734150, 93070977, 108094505},
								 {13772978, 29017588, 73764164, 169933419, 202917121},
								 {27289778, 57222646, 143824182, 323658301, 392561685}}

	elemSizes := []int{512,1024,2048}
	dbSizes := []int{150, 250, 500, 1000,1400}//,1730} //bigger db sizes tuned give problem: does not fit in uint16
	singlePassArr := make([][]float64,len(dbSizes)*len(elemSizes))
	//checklistArr := make([][]float64,len(dbSizes)*len(elemSizes))
	//cliSizeArr := make([][]int,len(dbSizes))
	numIter := 1
	
	for k,elemSize := range elemSizes {
		fmt.Printf("current elem Size: %d ",elemSize)

		
		for i,v := range dbSizes {

			dbSize := v*v//1<<i
			fmt.Printf("and current dbSize (singlepass) %d \n",dbSize)
			ddb := MakeDynamicDB(dbSize, elemSize)


			iterationArr := make([]float64,7)

			iterationArr[0] = float64(elemSize)
			iterationArr[1] = float64(dbSize)
			start := time.Now()
			elapsed := 0.0
			spSetSize := GetSetSizeByCliSizeDynamicSinglePass(dbSize,elemSize,checklistCliSizes[k][i])
			fmt.Printf("set size: %d\n",spSetSize)
			client := NewPIRReaderSetSize(RandSource(), Server(&ddb), Server(&ddb),spSetSize)
			err := client.Init(SinglePass)

			for j:= 0; j <numIter; j++ {
				start = time.Now()
				err = client.Init(SinglePass)
				elapsed += time.Since(start).Seconds()
				assert.NilError(t, err)
			}
			elapsedAvg := elapsed/float64(numIter)

			iterationArr[2] = elapsedAvg


			updateIdxs, updateRows := MakeUpdateRows(500, dbSize,elemSize)

			err=ddb.EditRows(updateIdxs,updateRows)
			//fmt.Println(ddb.ops)
			assert.NilError(t, err)
			elapsed = 0.0
			start = time.Now()
			err = client.Update()
			elapsed = time.Since(start).Seconds()
			fmt.Printf("update time: %f \n", elapsed)
			assert.NilError(t, err)

			readCtr := 0.0
			elapsed = 0.0
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, ddb.Row(j))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j+1 < dbSize {
					start = time.Now()
					val,err:=client.Read(j+1)
					elapsed += time.Since(start).Seconds()
					readCtr += 1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, ddb.Row(j+1))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, ddb.Row(j))
				}
			}
			amTime := elapsed/readCtr
			iterationArr[3] = amTime

			iterationArr[4] = float64(client.ClientSize())
			offBW, onBW := getSinglePassBandwidth(dbSize,elemSize,spSetSize)
			iterationArr[5] = float64(offBW)
			iterationArr[6] = float64(onBW)

			singlePassArr[k*len(dbSizes) + i] = iterationArr
		}

		
	}
	fmt.Println(singlePassArr)

}



func TestUpdatabaleSinglePass(t *testing.T) {
	if testing.Short() {
  		t.Skip("skipping testing in short mode")
	}
	elemSize := 2048
	dbSize := 1 << 20
	ddb := MakeDynamicDB(dbSize, elemSize)
	client := NewPIRReader(RandSource(), Server(&ddb), Server(&ddb))
	err := client.Init(SinglePass)
	assert.NilError(t, err)
	val,err :=client.Read(3)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, ddb.Row(3))

	numUpdates := 5
	numAdds := 5
	updateIdxs, updateRows := MakeUpdateRows(numUpdates, dbSize,elemSize)

	err=ddb.EditRows(updateIdxs,updateRows)
	//fmt.Println(ddb.ops)
	assert.NilError(t, err)
	_,addedRows := MakeUpdateRows(numAdds, dbSize, elemSize)
	err = ddb.AppendRows(addedRows)
	assert.NilError(t, err)

	err = client.Update()
	assert.NilError(t, err)

	for i := 0; i < numUpdates; i++ {

		val,err = client.Read(updateIdxs[i])
		assert.NilError(t, err)
		assert.DeepEqual(t, val, ddb.Row(updateIdxs[i]))
	}


	for i := 0; i < numAdds; i++ {

		val,err = client.Read(dbSize + i)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, ddb.Row(dbSize+i))
	}
	dbSize += numAdds
	
	updateIdxs, updateRows = MakeUpdateRows(numUpdates, dbSize,elemSize)
	err=ddb.EditRows(updateIdxs,updateRows)
	//fmt.Println(ddb.ops)
	assert.NilError(t, err)
	_,addedRows = MakeUpdateRows(numAdds, dbSize, elemSize)
	err = ddb.AppendRows(addedRows)


	err = client.Update()
	assert.NilError(t, err)

	for i := 0; i < numUpdates; i++ {
		val,err = client.Read(updateIdxs[i])
		assert.NilError(t, err)
		assert.DeepEqual(t, val, ddb.Row(updateIdxs[i]))
	}

	for i := 0; i < numAdds; i++ {
		val,err = client.Read(dbSize + i)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, ddb.Row(dbSize+i))
	}
	dbSize += numAdds
}


func TestFineGrainSinglePass(t *testing.T) {
	// if testing.Short() {
 //  		t.Skip("skipping testing in short mode")
	// }
	elemSize := 32
	dbSize := 1730*1730;
	db := MakeDB(dbSize, elemSize)
	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(Punc)
	elapsed := time.Since(start)
	fmt.Printf("checklist init time: %s \n", elapsed)
	assert.NilError(t, err)
	checklistCliSize := client.ClientSize()
	fmt.Printf("checklist cliSize: %d\n",checklistCliSize)

	start = time.Now()
	val, err := client.Read(1000000)
	elapsed = time.Since(start)
	fmt.Printf("checklist query time: %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(1000000))

	spSetSize := GetSetSizeByCliSizeStaticSinglePass(dbSize,elemSize,checklistCliSize)
	//spSetSize = 20
	client = NewPIRReaderSetSize(RandSource(), Server(db), Server(db), spSetSize)
	start = time.Now()
	err = client.Init(SinglePass)
	elapsed = time.Since(start)
	fmt.Printf("sp init time: %s \n", elapsed)
	fmt.Printf("sp cli size: %d \n", client.ClientSize())

	assert.NilError(t, err)
	start = time.Now()
	val, err = client.Read(1000000)
	elapsed = time.Since(start)
	fmt.Printf("sp query time: %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(1000000))
	fmt.Printf("sp set size: %d \n",spSetSize)
	fmt.Printf("sp cli size %d \n", client.ClientSize())

	_, chkOnBW := getChecklistBandwidth(dbSize,elemSize)
	_, spOnBW := getSinglePassBandwidth(dbSize,elemSize,spSetSize)
	fmt.Printf("checklist online bandwidth: %d, singlepass online bandwidth: %d \n", chkOnBW,spOnBW)

	//clear db to not take up memory:
	db = MakeDB(10,32)
	fmt.Println("--------Dynamic----------")
	dynamicSpSetSize := GetSetSizeByCliSizeDynamicSinglePass(dbSize,elemSize,25000000)
	ddb := MakeDynamicDB(dbSize,elemSize)
	client = NewPIRReaderSetSize(RandSource(), Server(&ddb), Server(&ddb),dynamicSpSetSize)
	start = time.Now()
	err = client.Init(SinglePass)
	elapsed = time.Since(start)
	fmt.Printf("dynamic sp init time: %s \n", elapsed)
	assert.NilError(t, err)
	

	numUpdates := 500
	//numAdds := 5
	updateIdxs, updateRows := MakeUpdateRows(numUpdates, dbSize,elemSize)

	err=ddb.EditRows(updateIdxs,updateRows)
	//fmt.Println(ddb.ops)
	assert.NilError(t, err)
	//_,addedRows := MakeUpdateRows(numAdds, dbSize, elemSize)
	//err = ddb.AppendRows(addedRows)
	//assert.NilError(t, err)
	start = time.Now()
	err = client.Update()
	elapsed = time.Since(start)
	assert.NilError(t, err)
	fmt.Printf("dyn sp total update time (500 updates): %s \n", elapsed)


	start = time.Now()
	val, err = client.Read(1000000)
	elapsed = time.Since(start)
	fmt.Printf("dyn sp query time: %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, ddb.Row(1000000))
	fmt.Printf("dyn sp set size: %d \n",dynamicSpSetSize)
	fmt.Printf("dyn sp cli size %d \n", client.ClientSize())
	_,spOnBW2 := getSinglePassBandwidth(dbSize,elemSize,dynamicSpSetSize)
	fmt.Printf(" online bandwidth dynamic sp: %d \n",spOnBW2)
}


func TestChecklistAndSinglePassSameCliSize(t *testing.T) {
	if testing.Short() {
  		t.Skip("skipping testing in short mode")
	}
	elemSizes := []int{512,1024,2048}
	dbSizes := []int{150, 250, 500, 1000,1400}//,1730} //bigger db sizes tuned give problem: does not fit in uint16
	singlePassArr := make([][]float64,len(dbSizes)*len(elemSizes))
	checklistArr := make([][]float64,len(dbSizes)*len(elemSizes))
	//cliSizeArr := make([][]int,len(dbSizes))
	numIter := 1
	
	for k,elemSize := range elemSizes {
		

		checklistCliSizes := make([]int, len(dbSizes))
		for i,v := range dbSizes {

			dbSize := v*v//1<<i
			
			db := MakeDB(dbSize, elemSize)
			iterationArr := make([]float64,7)

			iterationArr[0] = float64(elemSize)
			iterationArr[1] = float64(dbSize)
			start := time.Now()
			elapsed := 0.0
			client := NewPIRReader(RandSource(), Server(db), Server(db))
		
			
			for j:= 0; j <numIter; j++ {
				start = time.Now()
				err := client.Init(Punc)
				elapsed += time.Since(start).Seconds()
				assert.NilError(t, err)
			}
			elapsedAvg := elapsed/float64(numIter)

			iterationArr[2] = elapsedAvg
			checklistCliSizes[i] = client.ClientSize()
			
			readCtr := 0.0
			elapsed = 0.0
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
					if (dbSize == 1500*1500 && elemSize == 2048) {
						fmt.Printf("curr j = %d \n",j)
					}
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j+1 < dbSize {
					start = time.Now()
					val,err:=client.Read(j+1)
					elapsed += time.Since(start).Seconds()
					readCtr += 1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j+1))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
				}
			}
			amTime := elapsed/readCtr
			iterationArr[3] = amTime

			iterationArr[4] = float64(client.ClientSize())
			offBW, onBW := getChecklistBandwidth(dbSize,elemSize)
			iterationArr[5] = float64(offBW)
			iterationArr[6] = float64(onBW)

			checklistArr[k*len(dbSizes) + i] = iterationArr
			fmt.Println("checklist:")
			fmt.Println(iterationArr)
		}

		for i,v := range dbSizes {

			dbSize := v*v//1<<i
			
			db := MakeDB(dbSize, elemSize)
			iterationArr := make([]float64,7)

			iterationArr[0] = float64(elemSize)
			iterationArr[1] = float64(dbSize)
			start := time.Now()
			elapsed := 0.0
			spSetSize := GetSetSizeByCliSizeStaticSinglePass(dbSize,elemSize,checklistCliSizes[i])
			fmt.Printf("spSetSize: %d \n",spSetSize)
			client := NewPIRReaderSetSize(RandSource(), Server(db), Server(db),spSetSize)
			
			for j:= 0; j <numIter; j++ {
				start = time.Now()
				err := client.Init(SinglePass)
				elapsed += time.Since(start).Seconds()
				assert.NilError(t, err)
			}
			elapsedAvg := elapsed/float64(numIter)

			iterationArr[2] = elapsedAvg
			
			readCtr := 0.0
			elapsed = 0.0
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j+1 < dbSize {
					start = time.Now()
					val,err:=client.Read(j+1)
					elapsed += time.Since(start).Seconds()
					readCtr += 1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j+1))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
				}
			}
			amTime := elapsed/readCtr
			iterationArr[3] = amTime

			iterationArr[4] = float64(client.ClientSize())
			offBW, onBW := getSinglePassBandwidth(dbSize,elemSize,spSetSize)
			iterationArr[5] = float64(offBW)
			iterationArr[6] = float64(onBW)

			singlePassArr[k*len(dbSizes) + i] = iterationArr
			fmt.Println("single pass:")
			fmt.Println(iterationArr)

		}

	}
	
	fmt.Printf("checklist benchmarks same cliSize: \n")
	fmt.Println(checklistArr)
	fmt.Printf("SinglePass benchmarks same cliSize: \n")
	fmt.Println(singlePassArr)


}



func TestChecklistAndSinglePassSameSetSize(t *testing.T) {
	if testing.Short() {
  		t.Skip("skipping testing in short mode")
	}
	elemSizes := []int{512,1024,2048}//,3072}
	dbSizes := []int{150, 250, 500, 1000,1400}//, 2000}//,2000}
	singlePassArr := make([][]float64,len(dbSizes)*len(elemSizes))
	//checklistArr := make([][]float64,len(dbSizes)*len(elemSizes))
	//cliSizeArr := make([][]int,len(dbSizes))
	numIter := 1
	
	for k,elemSize := range elemSizes {

	// 	for i,v := range dbSizes {
	// 		dbSize := v*v//1<<i
	// 		db := MakeDB(dbSize, elemSize)
	// 		iterationArr := make([]float64,7)

	// 		iterationArr[0] = float64(elemSize)
	// 		iterationArr[1] = float64(dbSize)
	// 		start := time.Now()
	// 		elapsed := 0.0
	// 		client := NewPIRReader(RandSource(), Server(db), Server(db))
		
			
	// 		for j:= 0; j <numIter; j++ {
	// 			start = time.Now()
	// 			err := client.Init(Punc)
	// 			elapsed += time.Since(start).Seconds()
	// 			assert.NilError(t, err)
	// 		}
	// 		elapsedAvg := elapsed/float64(numIter)

	// 		iterationArr[2] = elapsedAvg

	// 		readCtr := 0.0
	// 		elapsed = 0.0
	// 		for j := 0; j < 202000; j+= 101 {
	// 			if j < dbSize {
	// 				start = time.Now()
	// 				val, err := client.Read(j)
	// 				elapsed += time.Since(start).Seconds()
	// 				readCtr +=1
	// 				assert.NilError(t, err)
	// 				assert.DeepEqual(t, val, db.Row(j))
	// 			}
	// 		}
	// 		for j := 0; j < 202000; j+= 101 {
	// 			if j+1 < dbSize {
	// 				start = time.Now()
	// 				val,err:=client.Read(j+1)
	// 				elapsed += time.Since(start).Seconds()
	// 				readCtr += 1
	// 				assert.NilError(t, err)
	// 				assert.DeepEqual(t, val, db.Row(j+1))
	// 			}
	// 		}
	// 		for j := 0; j < 202000; j+= 101 {
	// 			if j < dbSize {
	// 				start = time.Now()
	// 				val, err := client.Read(j)
	// 				elapsed += time.Since(start).Seconds()
	// 				readCtr +=1
	// 				assert.NilError(t, err)
	// 				assert.DeepEqual(t, val, db.Row(j))
	// 			}
	// 		}
	// 		amTime := elapsed/readCtr
	// 		iterationArr[3] = amTime


	// 		iterationArr[4] = float64(client.ClientSize())
	// 		offBW, onBW := getChecklistBandwidth(dbSize,elemSize)
	// 		iterationArr[5] = float64(offBW)
	// 		iterationArr[6] = float64(onBW)


	// 		checklistArr[k*len(dbSizes) + i] = iterationArr
	// 	}

		for i,v := range dbSizes {
			dbSize := v*v//1<<i
			db := MakeDB(dbSize, elemSize)
			iterationArr := make([]float64,7)

			iterationArr[0] = float64(elemSize)
			iterationArr[1] = float64(dbSize)
			start := time.Now()
			elapsed := 0.0
			//spSetSize := GetSetSizeByCliSizeStaticSinglePass(dbSize,elemSize,checklistCliSizes[i])
			client := NewPIRReader(RandSource(), Server(db), Server(db))
		
			for j:= 0; j <numIter; j++ {
				start = time.Now()
				err := client.Init(SinglePass)
				elapsed += time.Since(start).Seconds()
				assert.NilError(t, err)
			}
			elapsedAvg := elapsed/float64(numIter)

			iterationArr[2] = elapsedAvg

			readCtr := 0.0
			elapsed = 0.0
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j+1 < dbSize {
					start = time.Now()
					val,err:=client.Read(j+1)
					elapsed += time.Since(start).Seconds()
					readCtr += 1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j+1))
				}
			}
			for j := 0; j < 202000; j+= 101 {
				if j < dbSize {
					start = time.Now()
					val, err := client.Read(j)
					elapsed += time.Since(start).Seconds()
					readCtr +=1
					assert.NilError(t, err)
					assert.DeepEqual(t, val, db.Row(j))
				}
			}
			amTime := elapsed/readCtr
			iterationArr[3] = amTime

			iterationArr[4] = float64(client.ClientSize())
			offBW, onBW := getSinglePassBandwidth(dbSize,elemSize,0)
			iterationArr[5] = float64(offBW)
			iterationArr[6] = float64(onBW)

			singlePassArr[k*len(dbSizes) + i] = iterationArr
		}


	}
	
	//fmt.Printf("checklist benchmarks same set size (fixed sqrt(n): \n")
	//fmt.Println(checklistArr)
	fmt.Printf("SinglePass benchmarks same set size (fixed sqrt(n): \n")
	fmt.Println(singlePassArr)


}


func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 100)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Punc)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))

	// Test refreshing by reading the same item again
	val, err = client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))

}

func TestMatrix(t *testing.T) {
	db := MakeDB(10000, 4)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))
}

func TestDPF(t *testing.T) {
	

	elemSizes := []int{32}//[]int{512,1024,2048}//,3072}
	dbSizes := []int{1730}//[]int{150, 250, 500, 1000,1400}
	randy := RandSource()
	for _,e := range(elemSizes) {
		for _,v := range(dbSizes) {
			dbSize := v*v
			db := MakeDB(dbSize, e)
			client := NewPIRReader(RandSource(), Server(db), Server(db))
			err := client.Init(DPF)
			assert.NilError(t, err)
			elapsed := 0.0
			numQueries := 100
			for i := 0; i < numQueries; i++ {
				currQ := randy.Intn(dbSize)
				start := time.Now()
				val, err := client.Read(currQ)
				elapsed += time.Since(start).Seconds()
				assert.NilError(t, err)
				assert.DeepEqual(t, val, db.Row(currQ))
			}
			
			fmt.Printf("DPF PIR: elem Size: %d, DB Size: %d, query time: %f \n",e,dbSize,elapsed/float64(numQueries))

		}
	}
	

	// client := NewPIRReader(RandSource(), Server(db), Server(db))

	// err := client.Init(DPF)
	// assert.NilError(t, err)

	// val, err := client.Read(128)
	// assert.NilError(t, err)
	// assert.DeepEqual(t, val, db.Row(128))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[i%len(letterBytes)]
	}
	return string(b)
}

func TestSample(t *testing.T) {
	client := puncClient{randSource: RandSource()}
	assert.Equal(t, 1, client.sample(10, 0, 10))
	assert.Equal(t, 2, client.sample(0, 10, 10))
	assert.Equal(t, 0, client.sample(0, 0, 10))
	count := make([]int, 3)
	for i := 0; i < 1000; i++ {
		count[client.sample(10, 10, 30)]++
	}
	for _, c := range count {
		assert.Check(t, c < 380)
		assert.Check(t, c > 280)
	}
}
