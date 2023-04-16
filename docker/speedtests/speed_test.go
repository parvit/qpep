package speedtests

import (
	"flag"
	"github.com/stretchr/testify/suite"
	"testing"
)

var testTimeout = flag.Int("test_timeout", 60, "execution timeout seconds")
var connections = flag.Int("connections_num", 1, "simultaneous tcp connections to make to the server")

func TestSpeedTestsConfigSuite(t *testing.T) {
	t.Log(*testTimeout)
	t.Log(*connections)

	var q SpeedTestsConfigSuite
	suite.Run(t, &q)
}

type SpeedTestsConfigSuite struct {
	suite.Suite
}
