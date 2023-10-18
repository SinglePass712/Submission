package pir

import (
	//"errors"
	"fmt"
	"io"
	"log"
	"math"
	//"time"
	"math/rand"
	"checklist/psetggm"
)

type SinglePassClient struct {
	nRows  int
	RowLen int

	setSize int

	hints []Row
	nHints int

	randSource         *rand.Rand

	//names are not incredibly descriptive of actual roles, left as is from checklist impl
	//idxToSetIdx stores permutation
	//setIdxToIdx stores inverse permutation (for updates)
	//idxToSetIdx[i][j] = p_i (j)
	//setIdxToIdx[i][j] = p_i^(-1) (j)
	//first array indexes permutation, second indexes either the perm or its inverse
	//note setIdxToIdx is only necessary/helpful in the updatable case

	idxToSetIdx [][]uint32
	setIdxToIdx [][]uint32

	currVersion int
}

type SinglePassHintReq struct {
	RandSeed  PRGKey
	SetSize   int
}

type SinglePassHintResp struct {
	NRows        int
	RowLen       int
	SetSize      int
	RandSeed     PRGKey
	Hints        []Row
	NHints       int
	Permutations []uint32
	InversePermutations []uint32
	Version      int
}

func NewSinglePassHintReq(randSource *rand.Rand, setSize int) *SinglePassHintReq {
	req := &SinglePassHintReq{
		RandSeed:  PRGKey{},
		SetSize:   setSize,
	}
	_, err := io.ReadFull(randSource, req.RandSeed[:])
	if err != nil {
		log.Fatalf("Failed to initialize random seed: %s", err)
	}
	return req
}

func (req *SinglePassHintReq) Process(db StaticDB) (HintResp, error) {

	
	
	setSize := int(math.Sqrt(float64(db.NumRows)))//int(math.Sqrt(float64(db.NumRows)))
	if req.SetSize != 0 {
		setSize = req.SetSize
	}

	nHints := int(db.NumRows/setSize) //assuming this divides for now

	hints := make([]Row, nHints)
	hintsBuf := make([]byte, nHints*db.RowLen)



	//TODO: need to also pass in permutation array so i can point to it without having to rerun this
	//(to not have to sample them twice. In practice the client could also do it, concurrently to the server
	// or before the server)

	//TODO: CHECK: should I be passing pointers here or whole items

	permutations := make([]uint32, db.NumRows)
	inverse_permutations := make([]uint32, db.NumRows)
	//squash PRGKEY
	squash_rand := 0
	for i := 0; i < len(req.RandSeed); i++ {
		squash_rand += int(req.RandSeed[i])
	}
	 //start := time.Now()
	psetggm.SinglePassAnswer(db.FlatDb, db.NumRows, setSize, db.RowLen, hintsBuf, squash_rand, permutations, inverse_permutations)
	 //elapsed := time.Since(start)
	//fmt.Printf("SinglePassPIR: time for singlepassanswer, DB w/ %d elems: %s \n", db.NumRows, (elapsed))

	//maybe cut this off for efficiency, see impact
	for i := 0; i < nHints; i++ {
		hints[i] = Row(hintsBuf[db.RowLen*i : db.RowLen*(i+1)])
	}

	return &SinglePassHintResp{
		Hints:     hints,
		NHints:    nHints,
		NRows:     db.NumRows,
		RowLen:    db.RowLen,
		SetSize:   setSize,
		RandSeed:  req.RandSeed,
		Permutations: permutations,
		InversePermutations: inverse_permutations,
		Version: db.Version,
	}, nil
}


func (resp *SinglePassHintResp) InitClient(source *rand.Rand) Client {
	c := SinglePassClient{
		randSource: source,
		nRows:      resp.NRows,
		RowLen:     resp.RowLen,
		setSize:    resp.SetSize,
		hints:      resp.Hints,
		nHints:     resp.NHints,
		currVersion: resp.Version,
	}
	
	c.initState(resp.Permutations, resp.InversePermutations)
	//fmt.Println(resp.InversePermutations)
	return &c
}


func (c *SinglePassClient) initState(permutations []uint32, inverse_permutations []uint32) {
	//note: permutation size is c.nHints
	//note: numPermutations is c.setSize
	c.idxToSetIdx = make([][]uint32, c.setSize) //change to uint16
	c.setIdxToIdx = make([][]uint32, c.setSize)


	ind := 0
	i := 0
	for ind < len(permutations) {

		//copying arrray into another, hopefully this is extremely fast but can benchmark later to see
        c.idxToSetIdx[i] = permutations[ind: ind + c.nHints]
        c.setIdxToIdx[i] = inverse_permutations[ind: ind + c.nHints]
        ind += c.nHints
        i += 1

    }

}

// Sample uniformly from a range from 0 to maxRange



func (c *SinglePassClient) findIndex(index int) (index_0 int, index_1 int, pos int) {
	if index >= c.nRows {
		return -1,-1,-1
	}

	//1) split i into its prefix and suffix (i1,i2) (maybe do this outside of func and reanme)

	i0 := int(index/c.nHints)
	i1 := index - (i0*c.nHints)

	//2) iterate through c.idxToSetIdx[i1] until we find i2.
	// let j = number s.t c.idxToSetIdx[i1][j] = i2
	//(could also save inverse map but more storage)
	//(although asymptotically not more, can benchmark later if significant saving)

	
	//with c.setIdxToIdx can do as follows:
	return i0,i1,int(c.setIdxToIdx[i0][i1])


	//w/o c.setIdxToIdx must do:
	
	// permIndex := -1
	// for j:= 0; j < c.setSize; j++ {
	// 	if int(c.idxToSetIdx[i0][j]) == i1 {
	// 		permIndex = j
	// 	}
	// }
	//return i0,i1,permIndex
	

}

type SinglePassQueryReq struct {
	PRSet []uint32 //only second half of index since array is ordered
	SetSize int
}

type SinglePassQueryResp struct {
	Answer    []byte
}



func (c *SinglePassClient) Query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	i = MathMod(i, c.nRows)

	index_0, _, pos := c.findIndex(i)
	

	setOnline := make([]uint32, c.setSize)
	setOffline := make([]uint32, c.setSize)
	randSwaps := make([]uint32, c.setSize)


	for j:=0; j< c.setSize; j++ {
		setOnline[j] = c.idxToSetIdx[j][pos]

		randSwaps[j] = c.randomIdx(c.nHints)
		setOffline[j] = c.idxToSetIdx[j][randSwaps[j]]
	}
	setOnline[index_0] = c.idxToSetIdx[index_0][randSwaps[index_0]]



	return []QueryReq{
			&SinglePassQueryReq{PRSet: setOffline, SetSize: c.setSize},
			&SinglePassQueryReq{PRSet: setOnline, SetSize: c.setSize},
		},
		func(resps []interface{}) (Row, error) {
			queryResps := make([]*SinglePassQueryResp, len(resps))
			var ok bool
			for i, r := range resps {
				if queryResps[i], ok = r.(*SinglePassQueryResp); !ok {
					return nil, fmt.Errorf("Invalid response type: %T, expected: *SinglePassQueryResp", r)
				}
			}

			return c.reconstruct(queryResps, index_0, pos, randSwaps)
		}
}




//UNUSED: if we want to abstract map functionality can do after
func (c *SinglePassClient) refreshMap(index_0 int, pos int,intrandIdxs []uint32) { //change these to uint16?
	
	//to do: take set indicies, (hopefully what index they are within the map) and swap within the map
	return
}




func (q *SinglePassQueryReq) Process(db StaticDB) (interface{}, error) {
	
	//ACTUALLY here server does not do anything other than index db at requested indices and return all of those
	setSize := q.SetSize


	resp := SinglePassQueryResp{Answer: make([]byte, int(setSize) * db.RowLen)}

	offset := int(db.NumRows/setSize)
	
	//would it be faster to copy in C rather than in Go? - yes by some margin
	// for i:= 0; i < q.SetSize; i++ {
	// 	index := offset*i + int(q.PRSet[i])
	// 	elem := db.Row(index)
		
	// 	for j:= 0; j< db.RowLen; j++ {
	// 		resp.Answer[i*db.RowLen + j] = elem[j]
	// 	}
	// }
	currOffset := 0
	for i:= 0; i < setSize; i++ {
		psetggm.CopyIn(resp.Answer[i*db.RowLen:(i*db.RowLen)+db.RowLen],db.FlatDb, offset*i + int(q.PRSet[i]),db.RowLen)
		
		//psetggm.CopyIn(resp.Answer[i*db.RowLen:(i*db.RowLen)+db.RowLen],db.Row(offset*i + int(q.PRSet[i])),db.RowLen)
		//psetggm.CopyIn(resp.Answer,i,db.FlatDb, offset*i + int(q.PRSet[i]),db.RowLen)
		currOffset += offset
	}

	

	return &resp, nil
}

func (c *SinglePassClient) reconstruct(resp []*SinglePassQueryResp, index_0 int, pos int, randSwaps []uint32) (Row, error) { 
	//CHECK this idea of passing pointers vs actual values and check which is faster
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}


	elemSize := len(c.hints[0])



	//TODO: between resp[0] and resp[1], double check which is from offline server and which is from online server
	//for now, assuming resp[0] is from offline server and resp[1] is from online
	out := make(Row, elemSize)
	
	xorResp0 := make(Row, elemSize)
	xorResp1 := make(Row, elemSize)

	psetggm.XorBlocksTogether(resp[0].Answer, xorResp0, elemSize,c.setSize)
	psetggm.XorBlocksTogether(resp[1].Answer, xorResp1, elemSize, c.setSize)

	//Iif bandwidth is a bottleneck it would be good to be able to do this in a 'streaming' fashion to cut online time
	//as of now it is not clear exactly how to interface this

	

	upos := uint32(pos)

	//1) c.hints[pos] = xorResp0 (assuming xorResp0) is the offline response
	//2) out = out xor xorResp1 xor resp[1].Answer[index_0]

	//indexing is kind of obnoxious, can probably use rows to do this better?
	psetggm.FastXorInto(out,xorResp1,elemSize)
	psetggm.FastXorInto(out, c.hints[pos],elemSize)
	psetggm.FastXorInto(out, resp[1].Answer[index_0*elemSize:(index_0+1)*elemSize],elemSize)


	c.hints[pos] = xorResp0
	for i := 0; i < c.setSize; i++ {

		//maybe find faster way to xor - fixed, C++ xor performs ~5x faster

		//3) c.hints[randSwaps[i]] = c.hints[randSwaps[i]] XOR resp[0].Answer[i] XOR resp[1].Answer[i]
		psetggm.FastXorInto(c.hints[randSwaps[i]], resp[0].Answer[i*elemSize:(i+1)*elemSize], elemSize)
		psetggm.FastXorInto(c.hints[randSwaps[i]], resp[1].Answer[i*elemSize:(i+1)*elemSize], elemSize)
		//4) 
		temp1 := c.idxToSetIdx[i][pos]
		//can remove temp2 if not updatable
		temp2 := c.idxToSetIdx[i][randSwaps[i]]
		//5) 
		c.idxToSetIdx[i][pos] = c.idxToSetIdx[i][randSwaps[i]]
		//6) 
		c.idxToSetIdx[i][randSwaps[i]] = temp1
		//7)
		//for updatable: need to update new datastructure setIdxToIdx
		c.setIdxToIdx[i][temp1] = randSwaps[i]
		c.setIdxToIdx[i][temp2] = upos

	}
	//fix xoring once more than necessary
	psetggm.FastXorInto(c.hints[randSwaps[index_0]],resp[1].Answer[index_0*elemSize:(index_0+1)*elemSize],elemSize)
	psetggm.FastXorInto(c.hints[randSwaps[index_0]], out,elemSize)
	return Row(out), nil
}

// Sample a random element within range.
func (c *SinglePassClient) randomIdx(rangeMax int) uint32{
	// TODO: If this is slow, use a more clever way to
	// pick the random element.

	return uint32(c.randSource.Intn(rangeMax))

}



// Sample a random element of the set that is not equal to `idx`.
func (c *SinglePassClient) randomIdxExcept(rangeMax int, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := c.randSource.Intn(rangeMax)
		if val != idx {
			return val
		}
	}
}

func (c *SinglePassClient) StateSize() (bitsPerKey, fixedBytes int) {
	if c.currVersion > 0  {
		return 0,int(len(c.hints) * c.RowLen) + int(float64(c.nRows)* (math.Log2(float64(c.nHints))/8)*2)
	}
	return 0,int(len(c.hints) * c.RowLen) + int(float64(c.nRows)* (math.Log2(float64(c.nHints))/8))
}


///Updatable ///////////




type SinglePassUpdateReq struct {
	currVersion int
}

type SinglePassUpdateResp struct {
	newVersion int
	updates []UpdateOp
}

func (c *SinglePassClient) Update() (UpdateReq, UpdateFunc) {
	updateReq := SinglePassUpdateReq{c.currVersion}

	return &updateReq,	

		func(resp interface{}) (error) {
			
			if rResp, ok := resp.(*SinglePassUpdateResp); ok {
    			return c.updateState(rResp)
			} else {
   				return fmt.Errorf("Invalid response type")
			}
		}
}

func (updateReq *SinglePassUpdateReq) GetUpdates(ddb DynamicDB) (interface{}, error) {
	//client too old, server deleted old versioning, cannot refresh -> must re-preprocess
	if updateReq.currVersion < ddb.opInitVersion {
		return nil,fmt.Errorf("Client version too old, client must redo preprocessing from scratch")
	}
	// no updates
	if (ddb.ops == nil) || (len(ddb.ops) == 0) {
		return make([]UpdateOp, 0), nil
	}


	opsArr := ddb.ops[(updateReq.currVersion-ddb.opInitVersion):]
	//fmt.Println(opsArr)
	return &SinglePassUpdateResp{ddb.currVersion, opsArr}, nil //remember to return &resp not resp

}

func (c *SinglePassClient) updateState(updateResp *SinglePassUpdateResp) (error) {


	for _,op := range updateResp.updates {

		if op.Index >= c.nRows {

			c.nRows += c.nHints
			c.setSize+=1
			//sample new permutation and inverse,
			//append to end of current permutation and inverse
			//amortized O(1) time.. can deamortize with some coordination with server
			newPerm := make([]uint32, c.nHints)
			newInvPerm := make([]uint32, c.nHints)

			//TODO: use real randomness from rand instead of 0
			psetggm.SinglePermutation(0,newPerm,newInvPerm, uint32(c.nHints))

			//append newPerm, invPerm to idxToSetIdx and setIdxToIdx
			c.idxToSetIdx = append(c.idxToSetIdx, newPerm)
			c.setIdxToIdx = append(c.setIdxToIdx, newInvPerm)
		}

		//op.Index is something we have, just find hint it belongs to (O(1) time), and xor out old val, xor in new val

		_,_,pos := c.findIndex(op.Index)
		
		psetggm.FastXorInto(c.hints[pos], op.OldData,c.RowLen)
		psetggm.FastXorInto(c.hints[pos], op.NewData,c.RowLen)
	}
	c.currVersion = updateResp.newVersion

	return nil
}



////////////////unused/legacy
func (resp *SinglePassHintResp) NumRows() int {
	return resp.NRows
}
/////////////////////////////////////////////////////////////////////////////////////
//function not needed/used but kept for backward compat w/ checklist query interface
func (c *SinglePassClient) DummyQuery() []QueryReq {
	q := SinglePassQueryReq{PRSet: make([]uint32,1), SetSize: 0}
	return []QueryReq{&q,&q}
}
