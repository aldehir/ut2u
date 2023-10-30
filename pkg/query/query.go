package query

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/aldehir/ut2u/pkg/encoding/ue2"
)

type Client struct {
	conn            *net.UDPConn
	notifyList      map[string][]chan<- queryResponse
	notifyListMutex sync.RWMutex
}

type ServerDetails struct {
	Info    ServerInfo
	Rules   []KeyValuePair
	Players []Player
}

type Header struct {
	Version int32
	Command Command
}

type ServerInfo struct {
	ServerID       int32
	IP             string
	Port           int32
	QueryPort      int32
	ServerName     ue2.ColorizedString
	MapName        ue2.ColorizedString
	GameType       ue2.ColorizedString
	CurrentPlayers int32
	MaxPlayers     int32
	Ping           int32
	Flags          int32
	SkillLevel     string
}

type KeyValuePair struct {
	Key   ue2.ColorizedString
	Value ue2.ColorizedString
}

type Player struct {
	Num     int32
	Name    ue2.ColorizedString
	Ping    int32
	Score   int32
	StatsID int32
}

type queryResponse struct {
	From    net.Addr
	Header  Header
	Payload []byte
}

type QueryOptions struct {
	// Amount of time to wait for server responses before timing out. If set to
	// low, it is possible to receive partial data as the server sends query
	// responses in 450 byte chunks.
	Timeout time.Duration
	Command Command
}

type QueryOption func(*QueryOptions)

type Command uint8

const (
	Ping Command = 1 << iota
	Rules
	Players
)

const (
	pingCommand            = 0
	rulesCommand           = 1
	playersCommand         = 2
	rulesAndPlayersCommand = 3
)

const Version = 128

var ErrNoResponse = errors.New("no response")
var ErrInvalidCommand = errors.New("invalid command")

func NewClient() (*Client, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:       conn,
		notifyList: make(map[string][]chan<- queryResponse),
	}
	go client.listen()

	return client, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func WithPlayers() QueryOption {
	return func(opts *QueryOptions) {
		opts.Command |= Players
	}
}

func WithRules() QueryOption {
	return func(opts *QueryOptions) {
		opts.Command |= Rules
	}
}

func WithTimeout(timeout time.Duration) QueryOption {
	return func(opts *QueryOptions) {
		opts.Timeout = timeout
	}
}

func (c *Client) Query(ctx context.Context, addr net.Addr, opts ...QueryOption) (ServerDetails, error) {
	options := QueryOptions{
		Timeout: 250 * time.Millisecond,
		Command: Ping,
	}

	for _, o := range opts {
		o(&options)
	}

	responses := make(chan queryResponse, 5)
	c.notify(addr, responses)

	if options.Command&Ping != 0 {
		c.sendCommand(addr, pingCommand)
	}

	if options.Command&Rules != 0 && options.Command&Players != 0 {
		c.sendCommand(addr, rulesAndPlayersCommand)
	} else {
		if options.Command&Rules != 0 {
			c.sendCommand(addr, rulesCommand)
		} else if options.Command&Players != 0 {
			c.sendCommand(addr, playersCommand)
		}
	}

	timer := time.NewTimer(options.Timeout)
	var details ServerDetails
	err := ErrNoResponse

loop:
	for {
		select {
		case resp := <-responses:
			err = enrichDetails(&details, resp)
			if err != nil {
				break loop
			}
		case <-ctx.Done():
			err = ctx.Err()
			break loop
		case <-timer.C:
			break loop
		}
	}

	timer.Stop()
	c.stop(addr, responses)
	close(responses)

	if err != nil {
		return ServerDetails{}, err
	}

	return details, nil
}

func enrichDetails(details *ServerDetails, resp queryResponse) error {
	switch resp.Header.Command {
	case pingCommand:
		return ue2.Unmarshal(resp.Payload, &details.Info)
	case rulesCommand:
		var kv KeyValuePair

		buf := bytes.NewBuffer(resp.Payload)
		decoder := ue2.NewDecoder(buf)

		for {
			err := decoder.Decode(&kv)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			details.Rules = append(details.Rules, kv)
		}
	case playersCommand:
		var player Player

		buf := bytes.NewBuffer(resp.Payload)
		decoder := ue2.NewDecoder(buf)

		for {
			err := decoder.Decode(&player)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			details.Players = append(details.Players, player)
		}
	default:
		log.Printf("received invalid command: %d", resp.Header.Command)
	}

	return nil
}

func (c *Client) listen() {
	packet := make([]byte, 1024)

	for {
		n, addr, err := c.conn.ReadFrom(packet)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("err: %v", err)
			}
			break
		}

		var header Header

		buf := bytes.NewBuffer(packet[:n])
		decoder := ue2.NewDecoder(buf)

		err = decoder.Decode(&header)
		if err != nil {
			log.Printf("err: could not decode header (%v)", err)
			continue
		}

		payload := make([]byte, buf.Len())
		copy(payload, buf.Bytes())
		c.dispatch(addr, header, payload)
	}
}

func (c *Client) notify(addr net.Addr, ch chan<- queryResponse) {
	c.notifyListMutex.Lock()
	defer c.notifyListMutex.Unlock()

	var channels []chan<- queryResponse
	if list, ok := c.notifyList[addr.String()]; ok {
		channels = list
	}

	c.notifyList[addr.String()] = append(channels, ch)
}

func (c *Client) stop(addr net.Addr, ch chan<- queryResponse) {
	c.notifyListMutex.Lock()
	defer c.notifyListMutex.Unlock()

	var channels []chan<- queryResponse
	if list, ok := c.notifyList[addr.String()]; ok {
		channels = list
	}

	for i := len(channels) - 1; i >= 0; i-- {
		if channels[i] == ch {
			channels[i] = channels[len(channels)-1]
			channels = channels[:len(channels)-1]
		}
	}

	c.notifyList[addr.String()] = channels
}

func (c *Client) dispatch(addr net.Addr, header Header, payload []byte) {
	c.notifyListMutex.RLock()
	defer c.notifyListMutex.RUnlock()

	if channels, ok := c.notifyList[addr.String()]; ok {
		for _, ch := range channels {
			ch <- queryResponse{addr, header, payload}
		}
	} else {
		log.Printf("no channels for %s", addr.String())
	}
}

func (c *Client) sendCommand(addr net.Addr, cmd Command) error {
	payload, err := ue2.Marshal(Header{
		Version: Version,
		Command: cmd,
	})

	if err != nil {
		return err
	}

	_, err = c.conn.WriteTo(payload, addr)
	if err != nil {
		return err
	}

	return nil
}
