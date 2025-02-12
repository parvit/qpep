/*
 * This file in shared package handles the converstion to and from bytes
 * of the qpep header, used to instantiate new outer connections through
 * the quic connections.
 */
package shared

import (
	"encoding/binary"
	"io"
	"net"
)

// QPEP_PREAMBLE_LENGTH Length in bytes of the qpep header preamble that indicates the ip versions of the addresses
// in the qpep header
const (
	QPEP_PREAMBLE_LENGTH = 2
	IPV4                 = 0x04
	IPV6                 = 0x06
	IPNULL               = 0x00
)

// QPepHeader The structure that conveys the information about the destination and source of the connection
// established from the qpep client to the qpep server (ie. the DestAddr indicates the intended external destination
// different from the server of the connection)
type QPepHeader struct {
	// Source address of the client
	SourceAddr *net.TCPAddr
	// Destination of the connection (to be established by the server)
	DestAddr *net.TCPAddr
}

// ToBytes method of QPepHeader struct converts the header instance to a bytes
// slice at most 42 bytes long that represents the header struct
func (header QPepHeader) ToBytes() []byte {
	sourceType := getNetworkTypeFromAddr(header.SourceAddr)
	destType := getNetworkTypeFromAddr(header.DestAddr)
	if sourceType == IPNULL || destType == IPNULL {
		return nil
	}

	var byteOutput = make([]byte, 0, 42) // at most 2+16+4+16+4 = 42
	byteOutput = append(byteOutput, sourceType)
	byteOutput = append(byteOutput, destType)

	byteOutput = append(byteOutput, ipToBytes(header.SourceAddr.IP, sourceType)...)
	byteOutput = append(byteOutput, portToBytes(header.SourceAddr.Port)...)

	byteOutput = append(byteOutput, ipToBytes(header.DestAddr.IP, destType)...)
	byteOutput = append(byteOutput, portToBytes(header.DestAddr.Port)...)

	return byteOutput
}

// ipToBytes converts the indicated ip address to the bytes slice of the
// type indicated by _addrType_ (suitable for network transmission)
func ipToBytes(addr net.IP, addrType byte) []byte {
	if addr == nil {
		return nil
	}
	if addrType == IPV4 {
		return addr.To4()
	}
	return addr.To16()
}

// portToBytes converts the numerical port indicated to its byte representation
// suitable for network transmission
func portToBytes(port int) []byte {
	if port < 0 || port > 65535 {
		return nil
	}
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(port))
	return result
}

// getNetworkTypeFromAddr extracts the network address type from the provided
// address IP struct
func getNetworkTypeFromAddr(addr *net.TCPAddr) byte {
	if addr == nil {
		return IPNULL
	}
	if addr.IP.To4() != nil {
		return IPV4
	}
	return IPV6
}

// QPepHeaderFromBytes method tries to extract a well-formed QPepHeader struct
// from the provided reader.
// Can return either:
// * A valid QPepHeader pointer and nil error
// * Nil header and ErrInvalidHeader, if preamble is not coherent
// * Nil header and ErrInvalidHeaderAddressType, if preamble values are invalid
// * Nil header and ErrInvalidHeaderDataLength, if the data length is not suitable for extracting the indicated header
func QPepHeaderFromBytes(stream io.Reader) (*QPepHeader, error) {
	preamble := make([]byte, QPEP_PREAMBLE_LENGTH)
	ipBytesNum, err := stream.Read(preamble)
	if ipBytesNum != 2 || err != nil {
		return nil, ErrInvalidHeader
	}

	var sourceIpEnd int
	if preamble[0] == IPV4 {
		sourceIpEnd = net.IPv4len
	} else if preamble[0] == IPV6 {
		sourceIpEnd = net.IPv6len
	} else {
		return nil, ErrInvalidHeaderAddressType
	}

	sourcePortEnd := sourceIpEnd + 2

	var destIpEnd int
	if preamble[1] == IPV4 {
		destIpEnd = sourcePortEnd + net.IPv4len
	} else if preamble[1] == IPV6 {
		destIpEnd = sourcePortEnd + net.IPv6len
	} else {
		return nil, ErrInvalidHeaderAddressType
	}
	destPortEnd := destIpEnd + 2

	byteInput := make([]byte, destPortEnd)
	readDataBytes, err := stream.Read(byteInput)
	if readDataBytes != destPortEnd || err != nil {
		return nil, ErrInvalidHeaderDataLength
	}

	srcIPAddr := net.IP(byteInput[0:sourceIpEnd])
	srcPort := int(binary.LittleEndian.Uint16(byteInput[sourceIpEnd:sourcePortEnd]))

	destIPAddr := net.IP(byteInput[sourcePortEnd:destIpEnd])
	destPort := int(binary.LittleEndian.Uint16(byteInput[destIpEnd:destPortEnd]))

	srcAddr := &net.TCPAddr{IP: srcIPAddr, Port: srcPort}
	dstAddr := &net.TCPAddr{IP: destIPAddr, Port: destPort}
	return &QPepHeader{SourceAddr: srcAddr, DestAddr: dstAddr}, nil
}
