package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"net"

	util "cse312.app/utility"
	"cse312.app/values"
)

//constants used for websocket protocol
const (
	ACTIVE_USERS = iota + 1 //1
	GAME                    //2
)

type WSFrame struct {
	Payload  []byte
	protocol uint
	opcode   uint
	fin      uint
	mask     uint
}

type WebSocket interface {
	GetChan() chan *WSFrame
	Send(*net.Conn, []byte) error
	Close(string)
}

//naming things is hard :(
type Websocket struct {
	c          *net.Conn
	socketChan chan *WSFrame
	username   string //TODO integrate this
}

func UpgradeConn(c net.Conn, key, username string) WebSocket {
	//parse websocket request and upgrade connection
	key += values.WebsocketGUID
	checksum := sha1.Sum([]byte(key))
	key = base64.StdEncoding.EncodeToString(checksum[:])
	util.SendResponse(c, []string{values.Headers["101"], values.Headers["connection"], values.Headers["upgrade"], "Sec-WebSocket-Accept: " + key + "\r\n"}, nil)

	//ActiveUsersMutex.Lock()
	//values.UpgradedConn[c] = true
	//ActiveUsersMutex.Unlock()
	addUser(username, &c)

	ws := &Websocket{c: &c, socketChan: make(chan *WSFrame)}
	go ws.HandleWebSocket()

	return ws
}

func (ws *Websocket) GetChan() chan *WSFrame {
	return ws.socketChan
}

func (ws *Websocket) Close(username string) {
	//values.UpgradedConnMutex.Lock()
	//delete(values.UpgradedConn, *ws.c)
	//values.UpgradedConnMutex.Unlock()
	removeUser(username, ws.c)

	(*ws.c).Close()
	//I'm not closing channel becuase it might cause a panic
	//it's better to let GO reclaim it later
}

func (ws *Websocket) HandleWebSocket() {
	for {
		frame, _ := getWebSocketFrame(*ws.c)
		if frame.opcode == 8 {
			frame = nil
		}

		//opcode might be 8 for tear-down
		//opcode will always be 1
		//if frame.opcode != 1 {
		//	//log.Panic("opcode != 1 for websocket frame")
		//}

		ws.socketChan <- frame
		if frame == nil {
			return
		}
		/*
			customProtocol := frame.payload[0]
			_ = customProtocol
			//integer meanings are defined above
			switch customProtocol {
			case ACTIVE_USERS:
				Active_Users(c)
			case GAME:
				Game(c)
			default:
			}
		*/
	}

}

func getWebSocketFrame(c net.Conn) (*WSFrame, error) {
	data := make([]byte, 2048)
	n, err := c.Read(data)

	//check for socket error
	if n == 0 {
		//simulate close frame
		return &WSFrame{opcode: 8}, err
	}
	data = data[:n]

	//https://tools.ietf.org/html/rfc6455#section-5.2
	//parse websocket
	payload, fin, opcode := parseWebSocketFrame(c, data, n)
	tmpFrame := []byte{}

	//read until we have a complete websocket frame
	for fin != uint(128) {
		data = make([]byte, 2048)
		n, err = c.Read(data)
		//check for socket error
		if n == 0 {
			//simulate close frame
			return &WSFrame{opcode: 8}, err
		}
		data = data[:n]
		tmpFrame, fin, _ = parseWebSocketFrame(c, data, n)
		payload = append(payload, tmpFrame...)
	}

	//this is incomplete, but sufficient for now
	//TODO advance payload by one integer to remove metadata
	return &WSFrame{Payload: payload, fin: fin, opcode: opcode}, nil
}

func parseWebSocketFrame(c net.Conn, data []byte, n int) ([]byte, uint, uint) {
	frame := &WSFrame{}
	fin := uint(data[0] & 240)   // 11110000
	opcode := uint(data[0] & 15) // 00001111
	if opcode == 8 {
		return nil, 1, 1
	}

	frame.fin = fin
	frame.opcode = opcode

	mask := uint(data[1] & 128)         // 10000000
	payloadLen := uint64(data[1] & 127) // 01111111
	pos := 2
	_, _, _ = fin, opcode, mask
	if payloadLen == 126 {
		payloadLen = uint64(binary.BigEndian.Uint16(data[2:4]))
		pos = 4
	} else if payloadLen == 127 {
		payloadLen = binary.BigEndian.Uint64(data[2:10])
		pos = 10
	}

	pos += 4
	payload := make([]byte, payloadLen)

	//read more data if necessary
	currentBytes := uint64(n - pos)
	for currentBytes < payloadLen {
		buff := make([]byte, uint64(payloadLen-currentBytes))
		n, _ := c.Read(buff)
		buff = buff[:n]
		data = append(data, buff...)
		//removes excess bytes - i.e. x00
		currentBytes += uint64(n)
	}

	copy(payload[:], data[pos:])
	for b := range payload {
		offset := 8 * (b % 4)
		if offset == 24 {
			offset--
		}

		maskingByte := data[pos-4+(b%4)]
		payload[b] = payload[b] ^ maskingByte
	}
	return payload, fin, opcode
}

func makeResponseFrame(payload []byte) []byte {
	responseFrame := make([]byte, 0)
	responseFrame = append(responseFrame, []byte{129}...) // 100000001 --> sets finbit and opcode
	if uint(len(payload)) < 126 {
		responseFrame = append(responseFrame, uint8(len(payload))) // mask bit is set to 0
	} else {
		responseFrame = append(responseFrame, []byte{126}...) //no mask bit and we need two extra bytes for len
		responsePayloadLen := make([]byte, 2)
		binary.BigEndian.PutUint16(responsePayloadLen, uint16(len(payload)))
		responseFrame = append(responseFrame, responsePayloadLen...)
	}
	responseFrame = append(responseFrame, payload...)
	return responseFrame
}

func (ws *Websocket) Send(c *net.Conn, payload []byte) error {
	frame := makeResponseFrame(payload)
	util.Sendall(*c, frame)
	return nil
}

func Active_Users(c net.Conn) {
	/*
		responseFrame := makeResponseFrame(frame.payload)
		_ = responseFrame
		//for conn := range upgradedConn {
		//	sendall(conn, responseFrame)
		//}
	*/
}

func Game(c net.Conn) {

}
