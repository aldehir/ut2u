package query

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/aldehir/ut2u/pkg/encoding/ue2"
)

type QueryHeader struct {
	Version int32
	Command Command
}

type QueryResponse struct {
	Header  QueryHeader
	Payload []byte
}

type ServerResponseLine struct {
	ServerID       int32
	IP             string
	Port           int32
	QueryPort      int32
	ServerName     string
	MapName        string
	GameType       string
	CurrentPlayers int32
	MaxPlayers     int32
	Ping           int32
	Flags          int32
	SkillLevel     string
}

type KeyValuePair struct {
	Key   string
	Value string
}

type ServerRules struct {
	Rules []KeyValuePair
}

type Command uint8

const (
	PingCommand   Command = 0
	RulesCommmand Command = 1
)

const Version = 128

func QueryPing(ctx context.Context, addr string) (ServerResponseLine, error) {
	conn, err := dialAndRequest(ctx, addr, PingCommand)
	if err != nil {
		return ServerResponseLine{}, err
	}

	resp, err := nextPacket(ctx, conn)
	if err != nil {
		return ServerResponseLine{}, err
	}

	buf := bytes.NewBuffer(resp.Payload)

	var result ServerResponseLine
	err = ue2.Decode(buf, &result)
	if err != nil {
		return ServerResponseLine{}, err
	}

	return result, nil
}

func QueryRules(ctx context.Context, addr string) (ServerRules, error) {
	conn, err := dialAndRequest(ctx, addr, RulesCommmand)
	if err != nil {
		return ServerRules{}, err
	}

	resp, err := nextPacket(ctx, conn)
	if err != nil {
		return ServerRules{}, err
	}

	var buf bytes.Buffer
	buf.Write(resp.Payload)

	var result ServerRules
	var kv KeyValuePair

	for {
		err = ue2.Decode(&buf, &kv)
		if err == io.EOF {
			break
		}

		fmt.Println(kv)
		result.Rules = append(result.Rules, kv)
	}

	return result, nil
}

func dialAndRequest(ctx context.Context, addr string, cmd Command) (*net.UDPConn, error) {
	conn, err := dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	err = sendCommand(conn, cmd)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func dial(ctx context.Context, addr string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendCommand(conn *net.UDPConn, cmd Command) error {
	var buf bytes.Buffer

	err := ue2.Encode(&buf, QueryHeader{Version: Version, Command: cmd})
	if err != nil {
		return err
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func nextPacket(ctx context.Context, conn *net.UDPConn) (QueryResponse, error) {
	var err error
	packet := make([]byte, 8192)

	n, err := conn.Read(packet)
	if err != nil {
		return QueryResponse{}, err
	}

	var resp QueryResponse

	buf := bytes.NewBuffer(packet[:n])

	err = ue2.Decode(buf, &resp.Header)
	if err != nil {
		return QueryResponse{}, err
	}

	b := buf.Bytes()
	resp.Payload = make([]byte, len(b))
	copy(resp.Payload, b)

	return resp, nil
}
