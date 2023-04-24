package server

import (
	"context"
	"fmt"
	"github.com/Project-Faster/quic-go"
	"github.com/parvit/qpep/api"
	"github.com/parvit/qpep/logger"
	"github.com/parvit/qpep/shared"
	"io"
	"net"
	"runtime/debug"
	"runtime/trace"
	"sync"
	"time"
)

// listenQuicSession handles accepting the sessions and the launches goroutines to actually serve them
func listenQuicSession() {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()
	for {
		quicSession, err := quicListener.Accept(context.Background())
		if err != nil {
			logger.Error("Unrecoverable error while accepting QUIC session: %s\n", err)
			return
		}
		go func() {
			tskKey := fmt.Sprintf("ListenStream:%v", quicListener)
			_, tskStream := trace.NewTask(context.Background(), tskKey)

			listenQuicConn(quicSession)

			tskStream.End()
		}()
	}
}

// listenQuicConn handles opened quic sessions and accepts connections in goroutines to actually serve them
func listenQuicConn(quicSession quic.Connection) {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()
	for {
		stream, err := quicSession.AcceptStream(context.Background())
		if err != nil {
			if err.Error() != "NO_ERROR: No recent network activity" {
				logger.Error("Unrecoverable error while accepting QUIC stream: %s\n", err)
			}
			return
		}
		go func() {
			tskKey := fmt.Sprintf("QuicStream:%v", stream.StreamID())
			_, tskStream := trace.NewTask(context.Background(), tskKey)

			logger.Debug(">> QUIC StreamID START: %d\n", stream.StreamID())
			handleQuicStream(stream)
			logger.Debug(">> QUIC StreamID END: %d\n", stream.StreamID())

			tskStream.End()
		}()
	}
}

// handleQuicStream handles a quic stream connection and bridges to the standard tcp for the common internet
func handleQuicStream(stream quic.Stream) {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()

	qpepHeader, err := shared.QPepHeaderFromBytes(stream)
	if err != nil {
		logger.Info("Unable to decode QPEP header: %s\n", err)
		_ = stream.Close()
		return
	}

	srcLimit, okSrc := shared.GetAddressSpeedLimit(qpepHeader.SourceAddr.IP, true)
	dstLimit, okDst := shared.GetAddressSpeedLimit(qpepHeader.DestAddr.IP, false)
	if (okSrc && srcLimit == 0) || (okDst && dstLimit == 0) {
		logger.Info("Server speed limits blocked the connection, src:%v(%v) dst:%v(%v)", srcLimit, okSrc, dstLimit, okDst)
		_ = stream.Close()
		return
	}

	// To support the server being behind a private NAT (external gateway address != local listening address)
	// we dial the listening address when the connection is directed at the non-local API server
	destAddress := qpepHeader.DestAddr.String()
	if qpepHeader.DestAddr.Port == ServerConfiguration.APIPort {
		destAddress = fmt.Sprintf("%s:%d", ServerConfiguration.ListenHost, ServerConfiguration.APIPort)
	} else if qpepHeader.Flags&shared.QPEP_LOCALSERVER_DESTINATION != 0 {
		logger.Info("Local connection to server")
		destAddress = fmt.Sprintf("127.0.0.1:%d", qpepHeader.DestAddr.Port)
	}

	tskKey := fmt.Sprintf("TCP-Dial:%v", destAddress)
	_, tsk := trace.NewTask(context.Background(), tskKey)
	logger.Debug(">> Opening TCP Conn to dest:%s, src:%s\n", destAddress, qpepHeader.SourceAddr)
	dial := &net.Dialer{
		LocalAddr:     &net.TCPAddr{IP: net.ParseIP(ServerConfiguration.ListenHost)},
		Timeout:       shared.GetScaledTimeout(1, time.Second),
		KeepAlive:     shared.GetScaledTimeout(3, time.Second),
		DualStack:     true,
		FallbackDelay: 10 * time.Millisecond,
	}
	tcpConn, err := dial.Dial("tcp", destAddress)
	tsk.End()
	if err != nil {
		logger.Error("Unable to open TCP connection from QPEP stream: %s\n", err)
		stream.Close()

		shared.ScaleUpTimeout()
		return
	}
	logger.Debug(">> Opened TCP Conn %s -> %s\n", qpepHeader.SourceAddr, destAddress)

	trackedAddress := qpepHeader.SourceAddr.IP.String()
	proxyAddress := tcpConn.LocalAddr().String()

	api.Statistics.IncrementCounter(1.0, api.TOTAL_CONNECTIONS)
	api.Statistics.IncrementCounter(1.0, api.PERF_CONN, trackedAddress)
	defer func() {
		api.Statistics.DecrementCounter(1.0, api.PERF_CONN, trackedAddress)
		api.Statistics.DecrementCounter(1.0, api.TOTAL_CONNECTIONS)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	var streamWait sync.WaitGroup
	streamWait.Add(2)

	go handleQuicToTcp(ctx, &streamWait, srcLimit, tcpConn, stream, proxyAddress, trackedAddress)
	go handleTcpToQuic(ctx, &streamWait, dstLimit, stream, tcpConn, trackedAddress)
	go func() {
		<-time.After(180 * time.Second)
		cancel()
	}()

	//we exit (and close the TCP connection) once both streams are done copying or timeout
	streamWait.Wait()

	stream.CancelRead(0)
	stream.CancelWrite(0)
	stream.Close()
	tcpConn.Close()

	logger.Debug(">> Closing TCP Conn %s->%s\n", proxyAddress, tcpConn.RemoteAddr().String())
}

const (
	BUFFER_SIZE = 512 * 1024
)

func handleQuicToTcp(ctx context.Context, streamWait *sync.WaitGroup, speedLimit int64,
	dst net.Conn, src quic.Stream, proxyAddress, trackedAddress string) {

	tskKey := fmt.Sprintf("Tcp->Quic:%v", src.StreamID())
	_, tsk := trace.NewTask(context.Background(), tskKey)
	defer func() {
		_ = recover()

		dst.Close()
		api.Statistics.DeleteMappedAddress(proxyAddress)
		tsk.End()
		streamWait.Done()
	}()

	api.Statistics.SetMappedAddress(proxyAddress, trackedAddress)

	setLinger(dst)

	var loopTimeout = 1 * time.Second
	var tempBuffer = make([]byte, BUFFER_SIZE)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var written int64 = 0
		var err error

		tm := time.Now().Add(loopTimeout)
		_ = src.SetReadDeadline(tm)
		_ = src.SetWriteDeadline(tm)
		_ = dst.SetReadDeadline(tm)
		_ = dst.SetWriteDeadline(tm)

		_, tsk := trace.NewTask(context.Background(), "copybuffer."+tskKey)
		if speedLimit > 0 {
			var start = time.Now()
			var limit = start.Add(loopTimeout)
			written, err = io.CopyN(dst, src, speedLimit)
			var end = limit.Sub(time.Now())

			//logger.Debug("q -> t: %d / %v", written, end.Nanoseconds())
			time.Sleep(end)

		} else {
			written, err = io.CopyBuffer(dst, io.LimitReader(src, BUFFER_SIZE), tempBuffer)
			logger.Debug("q -> t: %d", written)
		}
		tsk.End()

		go api.Statistics.IncrementCounter(float64(written), api.PERF_UP_COUNT, trackedAddress)

		if err != nil || written == 0 {
			if nErr, ok := err.(net.Error); ok && (nErr.Timeout() || nErr.Temporary()) {
				continue
			}
			//log.Printf("Error on Copy %s\n", err)
			//logger.Debug("finish q -> t")
			return
		}
	}
}

func handleTcpToQuic(ctx context.Context, streamWait *sync.WaitGroup, speedLimit int64,
	dst quic.Stream, src net.Conn, trackedAddress string) {

	tskKey := fmt.Sprintf("Tcp->Quic:%v", dst.StreamID())
	_, tsk := trace.NewTask(context.Background(), tskKey)
	defer func() {
		_ = recover()

		dst.Close()
		tsk.End()
		streamWait.Done()
	}()

	setLinger(src)

	var loopTimeout = 1 * time.Second

	var tempBuffer = make([]byte, BUFFER_SIZE)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var written int64 = 0
		var err error

		tm := time.Now().Add(loopTimeout)
		_ = src.SetReadDeadline(tm)
		_ = src.SetWriteDeadline(tm)
		_ = dst.SetReadDeadline(tm)
		_ = dst.SetWriteDeadline(tm)

		_, tsk := trace.NewTask(context.Background(), "copybuffer."+tskKey)
		if speedLimit > 0 {
			var start = time.Now()
			var limit = start.Add(loopTimeout)
			written, err = io.CopyN(dst, src, speedLimit)
			var end = limit.Sub(time.Now())

			//logger.Debug("t -> q: %d / %v", written, end.Nanoseconds())
			time.Sleep(end)

		} else {
			written, err = io.CopyBuffer(dst, io.LimitReader(src, BUFFER_SIZE), tempBuffer)
			logger.Debug("t -> q: %d", written)
		}
		tsk.End()

		go api.Statistics.IncrementCounter(float64(written), api.PERF_DW_COUNT, trackedAddress)

		if err != nil || written == 0 {
			if nErr, ok := err.(net.Error); ok && (nErr.Timeout() || nErr.Temporary()) {
				continue
			}
			//logger.Debug("finish t -> q")
			return
		}
	}
}

func setLinger(c net.Conn) {
	if conn, ok := c.(*net.TCPConn); ok {
		err1 := conn.SetLinger(1)
		logger.OnError(err1, "error on setLinger")
	}
}
