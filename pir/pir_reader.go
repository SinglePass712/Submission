package pir

import (
	"fmt"
	"math/rand"
)

//go:generate enumer -type=PirType
type PirType int

const (
	None PirType = iota
	Matrix
	Punc
	Perm
	DPF
	NonPrivate
	SinglePass
)

type Server interface {
	Hint(req HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *interface{}) error
	UpdateServer(updateIdxs []int, newRows []Row) error
	UpdateClient(u UpdateReq, resp *interface{}) error
}

func NewHintReq(source *rand.Rand, pirType PirType, setSize int) HintReq {
	switch pirType {
	case Matrix:
		return NewMatrixHintReq()
	case Punc:
		return NewPuncHintReq(source)
	case DPF:
		return NewDPFHintReq()
	case NonPrivate:
		return NewNonPrivateHintReq()
	case SinglePass:
		return NewSinglePassHintReq(source, setSize)
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}

type PIRReader interface {
	Init(pirType PirType) error
	Read(i int) (Row, error)
	Update() error
	ClientSize() int
}

type pirReader struct {
	impl       Client
	servers    [2]Server
	randSource *rand.Rand
	setSize    int
}

func NewPIRReader(source *rand.Rand, serverL, serverR Server) PIRReader {
	return &pirReader{servers: [2]Server{serverL, serverR}, randSource: source, setSize:0}
}
func NewPIRReaderSetSize(source *rand.Rand, serverL, serverR Server, setSize int) PIRReader {
	return &pirReader{servers: [2]Server{serverL, serverR}, randSource: source,setSize:setSize}
}
func (c *pirReader) Init(pirType PirType) error {
	req := NewHintReq(c.randSource, pirType,c.setSize)
	var hintResp HintResp
	if err := c.servers[Left].Hint(req, &hintResp); err != nil {
		return err
	}
	c.impl = hintResp.InitClient(c.randSource)
	return nil
}

func (c pirReader) Read(i int) (Row, error) {
	if c.impl == nil {
		return nil, fmt.Errorf("Did you forget to call Init?")
	}
	queryReq, reconstructFunc := c.impl.Query(i)
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %d", i)
	}
	responses := make([]interface{}, 2)
	err := c.servers[Left].Answer(queryReq[Left], &responses[Left])
	if err != nil {
		return nil, err
	}

	err = c.servers[Right].Answer(queryReq[Right], &responses[Right])
	if err != nil {
		return nil, err
	}
	return reconstructFunc(responses)
}

func (c *pirReader) Update() error {
	if c.impl == nil {
		return fmt.Errorf("Did not call init")
	}
	if singPassCli, ok := c.impl.(*SinglePassClient); ok {

		//can separate updates to server to different function within pirreader later
		//c.servers[Left].UpdateServer(upInds, upRows)
		//c.servers[Right].UpdateServer(upInds, upRows)


		updateReq, updateFunc := singPassCli.Update()

		response := make([]interface{},1)
		err := c.servers[Left].UpdateClient(updateReq, &response[0])
		if err != nil {
			return err
		}
		
		updateFunc(response[0])


	} else {
		return fmt.Errorf("Update can only be called for Single Pass Client.")
	}
	return nil
}

func (c *pirReader) ClientSize() int {
	_,val := c.impl.StateSize()
	return val
}

