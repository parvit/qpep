package speedtests

import (
	"flag"
	"fmt"
	"github.com/parvit/qpep/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

var targetURL = flag.String("target_url", "", "url to download")
var connections = flag.Int("connections_num", 1, "simultaneous tcp connections to make to the server")
var expectedSize = flag.Int("expect_mb", 10, "size in MBs of the target file")

func TestSpeedTestsConfigSuite(t *testing.T) {
	t.Log(*targetURL)
	t.Log(*connections)
	t.Log(*expectedSize)

	assert.True(t, *connections > 0)
	assert.True(t, len(*targetURL) > 0)
	assert.True(t, *expectedSize > 0)

	*expectedSize = 1024 * 1024 * *expectedSize

	var q SpeedTestsConfigSuite
	suite.Run(t, &q)
}

type SpeedTestsConfigSuite struct {
	suite.Suite
}

func (s *SpeedTestsConfigSuite) TestRun() {
	shared.GetSystemProxyEnabled()

	wg := &sync.WaitGroup{}
	wg.Add(*connections)

	f, err := os.Create("output.csv")
	assert.Nil(s.T(), err)
	defer func() {
		_ = f.Sync()
		_ = f.Close()
	}()

	lock := &sync.Mutex{}

	f.WriteString("timestamp,event,value\n")

	for index := 0; index < *connections; index++ {
		go func(id int) {
			defer func() {
				wg.Done()
			}()
			s.T().Logf("Starting executor #%d\n", id)

			client := getClientForAPI(nil)
			assert.NotNil(s.T(), client)
			assert.NotNil(s.T(), targetURL)

			s.T().Logf("GET request #%d", id)
			resp, err := client.Get(*targetURL)
			assert.Nil(s.T(), err)
			if err != nil {
				s.T().Logf("GET request failed #%d", id)
				return
			}
			defer resp.Body.Close()

			toRead := resp.ContentLength
			if toRead != int64(*expectedSize) {
				assert.Fail(s.T(), "No response / wrong response")
				return
			}

			var events = make([]string, 0, 256)

			var eventTag = fmt.Sprintf("conn-%d-speed", id)

			var totalBytesInTimeDelta int64 = 0
			var start = time.Now()
			var buff = make([]byte, 1024)

			for toRead > 0 {
				rd := io.LimitReader(resp.Body, 1024)
				read, err := rd.Read(buff)
				if err != nil && err != io.EOF {
					if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
						continue
					}
					s.T().Logf("err: %v", err)
					assert.Failf(s.T(), "failed", "%v", err)
					return
				}
				if read == 0 {
					<-time.After(10 * time.Millisecond)
					continue
				}

				totalBytesInTimeDelta += int64(read)
				toRead -= int64(read)
				//s.T().Logf("#%d read: %d, toRead: %d", id, totalBytesInTimeDelta, toRead)
				if time.Since(start) > 1*time.Second {
					start = time.Now()
					s.T().Logf("#%d bytes to read: %d", id, toRead)
					events = append(events, fmt.Sprintf("%s,%s,%d\n", start.Format(time.RFC3339Nano), eventTag, totalBytesInTimeDelta/1024))
					totalBytesInTimeDelta = 0
				}
			}
			if totalBytesInTimeDelta > 0 {
				start = time.Now()
				events = append(events, fmt.Sprintf("%s,%s,%d\n", start.Format(time.RFC3339Nano), eventTag, totalBytesInTimeDelta/1024))
			}

			assert.True(s.T(), toRead <= 0)

			s.T().Logf("#%d GET request done, dumping to CSV...", id)
			lock.Lock()
			defer lock.Unlock()
			for _, ev := range events {
				f.WriteString(ev)
			}
			s.T().Logf("#%d done", id)
		}(index)
	}

	wg.Wait()
}

func getClientForAPI(localAddr net.Addr) *http.Client {
	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				shared.UsingProxy, shared.ProxyAddress = shared.GetSystemProxyEnabled()
				if shared.UsingProxy {
					return shared.ProxyAddress, nil
				}
				return nil, nil
			},
			DialContext:     dialer.DialContext,
			MaxIdleConns:    0,
			IdleConnTimeout: 10 * time.Second,
			//TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
