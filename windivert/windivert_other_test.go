//go:build !windows

package windivert

import (
	"github.com/parvit/qpep/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestWinDivertSuite(t *testing.T) {
	var q WinDivertSuite
	suite.Run(t, &q)
}

type WinDivertSuite struct {
	suite.Suite
}

func (s *WinDivertSuite) AfterTest(_, _ string) {}

func (s *WinDivertSuite) BeforeTest(_, _ string) {}

func (s *WinDivertSuite) TestInitializeWinDivertEngine() {
	t := s.T()

	addr, _ := shared.GetDefaultLanListeningAddress("127.0.0.1", "")

	code := InitializeWinDivertEngine(
		addr, addr,
		shared.QPepConfig.GatewayAPIPort, 445,
		4, 0)

	assert.Equal(t, DIVERT_OK, code)
}

func (s *WinDivertSuite) TestInitializeWinDivertEngine_Fail() {
	t := s.T()

	addr, _ := shared.GetDefaultLanListeningAddress("127.0.0.1", "")

	code := InitializeWinDivertEngine(
		addr, addr,
		0, 0,
		4, 0)

	assert.Equal(t, DIVERT_OK, code) // ok because it's not implemented on linux
}

func (s *WinDivertSuite) TestCloseWinDivertEngine() {
	t := s.T()

	addr, _ := shared.GetDefaultLanListeningAddress("127.0.0.1", "")

	code := InitializeWinDivertEngine(
		addr, addr,
		shared.QPepConfig.GatewayAPIPort, 445,
		4, 0)

	assert.Equal(t, DIVERT_OK, code)

	code = CloseWinDivertEngine()
	assert.Equal(t, DIVERT_OK, code)
}

func (s *WinDivertSuite) TestEnableDiverterLogging() {
	EnableDiverterLogging(true)
	EnableDiverterLogging(false)

	// only for coverage, no actual effect can be checked
}

func (s *WinDivertSuite) TestGetConnectionStateDate_Closed() {
	code, srcPort, dstPort, srcAddress, dstAddress := GetConnectionStateData(9999)
	assert.Equal(s.T(), DIVERT_OK, code) // ok because it's not implemented on linux

	assert.Equal(s.T(), srcPort, -1)
	assert.Equal(s.T(), dstPort, -1)
	assert.Equal(s.T(), srcAddress, "")
	assert.Equal(s.T(), dstAddress, "")
}
