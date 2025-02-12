package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/parvit/qpep/logger"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/parvit/qpep/shared"
)

// getClientForAPI method returns a correctly configured http client to be able to
// contact the qpep APIs (dynamically adapting to either diverter mode or proxy mode)
func getClientForAPI(localAddr net.Addr) *http.Client {
	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				logger.Info("API Proxy: %v %v\n", shared.UsingProxy, shared.ProxyAddress)
				if shared.UsingProxy {
					return shared.ProxyAddress, nil
				}
				return nil, nil
			},
			DialContext:     dialer.DialContext,
			MaxIdleConns:    1,
			IdleConnTimeout: 10 * time.Second,
			//TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// doAPIRequest method encapsulates the common logic required to execute an api request
// for the qpep APIs, client value is supposed to have been obtained from getClientForAPI
func doAPIRequest(addr string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		logger.Error("%v\n", err)
		return nil, err
	}
	req.Header.Set("User-Agent", runtime.GOOS)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("%v\n", err)
		return nil, err
	}
	return resp, nil
}

// RequestEcho method wraps the entire API connection logic to obtain an object of type
// EchoResponse to check for connection to a qpep server.
func RequestEcho(localAddress, address string, apiPort int, toServer bool) *EchoResponse {
	prefix := API_PREFIX_CLIENT
	if toServer {
		prefix = API_PREFIX_SERVER
	}

	resolvedAddr, errAddr := net.ResolveTCPAddr("tcp", localAddress+":0")
	if errAddr != nil {
		logger.Error("%v\n", errAddr)
		return nil
	}

	clientInst := getClientForAPI(resolvedAddr)

	addr := fmt.Sprintf("http://%s:%d%s", address, apiPort, prefix+API_ECHO_PATH)
	resp, err := doAPIRequest(addr, clientInst)
	if err != nil {
		logger.Error("%v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(" BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	str.Grow(8192)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.Write(scanner.Bytes())
	}

	if scanner.Err() != nil {
		logger.Error("%v\n", scanner.Err())
		return nil
	}

	if shared.QPepConfig.Verbose {
		logger.Info("%s\n", str.String())
	}

	respData := &EchoResponse{}
	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		logger.Error("%v\n", jsonErr)
		return nil
	}

	Statistics.SetState(INFO_OTHER_VERSION, respData.ServerVersion)
	return respData
}

// RequestStatus method wraps the entire API connection logic to obtain an object of type
// StatusResponse to query the current status of a server
func RequestStatus(localAddress, gatewayAddress string, apiPort int, publicAddress string, toServer bool) *StatusResponse {
	prefix := API_PREFIX_CLIENT
	if toServer {
		prefix = API_PREFIX_SERVER
	}
	apiPath := strings.Replace(prefix+API_STATUS_PATH, ":addr", publicAddress, -1)
	addr := fmt.Sprintf("http://%s:%d%s", gatewayAddress, apiPort, apiPath)

	resolvedAddr, errAddr := net.ResolveTCPAddr("tcp", localAddress+":0")
	if errAddr != nil {
		logger.Error("%v\n", errAddr)
		return nil
	}

	clientInst := getClientForAPI(resolvedAddr)

	resp, err := doAPIRequest(addr, clientInst)
	if err != nil {
		logger.Error("%v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(" BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	str.Grow(8192)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.Write(scanner.Bytes())
	}

	if scanner.Err() != nil {
		logger.Error("6  %v\n", scanner.Err())
		return nil
	}

	if shared.QPepConfig.Verbose {
		logger.Info("%s\n", str.String())
	}

	respData := &StatusResponse{}
	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		logger.Error("7  %v\n", jsonErr)
		return nil
	}

	return respData
}

// RequestStatistics method wraps the entire API connection logic to obtain an object of type
// StatsInfoResponse to obtain info about an host from a qpep server.
func RequestStatistics(localAddress, gatewayAddress string, apiPort int, publicAddress string) *StatsInfoResponse {
	apiPath := strings.Replace(API_PREFIX_SERVER+API_STATS_DATA_SRV_PATH, ":addr", publicAddress, -1)
	addr := fmt.Sprintf("http://%s:%d%s", gatewayAddress, apiPort, apiPath)

	resolvedAddr, errAddr := net.ResolveTCPAddr("tcp", localAddress+":0")
	if errAddr != nil {
		logger.Error("%v\n", errAddr)
		return nil
	}

	clientInst := getClientForAPI(resolvedAddr)

	resp, err := doAPIRequest(addr, clientInst)
	if err != nil {
		logger.Error("%v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(" BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	str.Grow(8192)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.Write(scanner.Bytes())
	}

	if scanner.Err() != nil {
		logger.Error("9  %v\n", scanner.Err())
		return nil
	}

	if shared.QPepConfig.Verbose {
		logger.Info("%s\n", str.String())
	}

	respData := &StatsInfoResponse{}
	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		logger.Error("10  %v\n", jsonErr)
		return nil
	}

	return respData
}
