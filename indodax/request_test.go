package indodax

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestTestSuite struct {
	suite.Suite
}

func (suite *RequestTestSuite) SetupTest() {
	_ = InitIndodax("OTI7CAZS-0AXCSLY5-HNSRQS8C-TLBEMMGC-PBWFEVBB",
		"095d11ae06e57eb888962983fc75d680e9bd19c24dd6cda1b575d838c1629289831cf1c2f0e9c1a0")
}

func (suite *RequestTestSuite) TestGetDepth() {
	d := IndodaxInstance.GetDepth("eth_idr")
	if d.IsEmpty() {
		suite.T().Fail()
		return
	}
}

func (suite *RequestTestSuite) TestGetInfo() {
	d := IndodaxInstance.GetInfo()
	fmt.Println("this is d")
	fmt.Println(d)

	suite.T().Fail()
}

func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}
