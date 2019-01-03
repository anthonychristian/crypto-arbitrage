package indodax

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestTestSuite struct {
	suite.Suite
}

func (suite *RequestTestSuite) SetupTest() {
	_ = InitIndodax()
}

func (suite *RequestTestSuite) TestGetDepth() {
	d := IndodaxInstance.getDepth("eth_idr")
	if d.IsEmpty() {
		suite.T().Fail()
		return
	}
}

func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}
