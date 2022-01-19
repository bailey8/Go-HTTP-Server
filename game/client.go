package game

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"cse312.app/websocket"
)

type Game struct {
	socketList      map[int]*net.Conn
	socketListMutex *sync.Mutex
	playerList      map[int]*Player
	playerListMutex *sync.Mutex
	bulletList      map[string]*Bullet
	bulletListMutex *sync.Mutex
	ticker          *time.Ticker
	r               *rand.Rand
	userChan        chan *NewUser //send a chan over a chan!
	userList        map[*NewUser]*Player
	id              int
	connectedUsers  int
}

//needed for cspGame
type NewUser struct {
	number      string
	id          int
	userInChan  chan *Frame
	userOutChan chan []byte
}

func NewGame() *Game {
	g := &Game{}
	g.ticker = time.NewTicker(40 * time.Millisecond)
	g.socketList = make(map[int]*net.Conn)
	g.socketListMutex = &sync.Mutex{}
	g.playerList = make(map[int]*Player)
	g.bulletList = make(map[string]*Bullet)
	g.bulletListMutex = &sync.Mutex{}
	g.playerListMutex = &sync.Mutex{}
	g.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	g.userChan = make(chan *NewUser)
	g.userList = make(map[*NewUser]*Player)
	return g
}

func (g *Game) PlayGame(c net.Conn, key, username string) {
	ws := websocket.UpgradeConn(c, key, username)
	defer ws.Close(username)

	rand.Seed(time.Now().UnixNano())
	socketId := rand.Intn(100000000)
	//number := strconv.Itoa(rand.Intn(10))

	//handle new connection
	player := g.newPlayer(socketId, username)
	done := make(chan bool)
	g.addSocket(socketId, &c)

	go func() { //setInterval
		for {
			select {
			case <-done:
				return
				//can use time.After(40 * time.Millisecond), but there would roughly twice as many messages.
				//using the ticker is okay because a node will braodcast a message to all other nodes!
			case <-g.ticker.C: //playerlist is empty
				players := g.updateAllPlayers()
				bullets := g.updateAllBullets()
				msg := Positions{
					Action: "newPositions",
					PList:  players,
					BList:  bullets,
				}
				msgByte, err := json.Marshal(msg)
				if err != nil {
					log.Println("error:", err)
				}
				//log.Println(string(msgByte), players[0])
				g.socketListMutex.Lock()
				for _, s := range g.socketList {
					ws.Send(s, msgByte)
				}
				g.socketListMutex.Unlock()
			}
		}
	}()
	for {
		frame := <-ws.GetChan()
		if frame == nil {
			//assume client disconnected
			g.removeSocket(socketId)
			done <- true
			return
		}

		var parsedFrame Frame
		err := json.Unmarshal(frame.Payload, &parsedFrame)
		_ = err
		switch parsedFrame.Action {
		case "keyPress":
			if parsedFrame.InputId == "left" {
				player.pressingLeft = parsedFrame.State
			} else if parsedFrame.InputId == "right" {
				player.pressingRight = parsedFrame.State
			} else if parsedFrame.InputId == "up" {
				player.pressingUp = parsedFrame.State
			} else if parsedFrame.InputId == "down" {
				player.pressingDown = parsedFrame.State
			} else if parsedFrame.InputId == "attack" {
				player.mouseAngle = parsedFrame.Angle
				player.pressingAttack = parsedFrame.State
			} else if parsedFrame.InputId == "mouseAngle" {
				//player.mouseAngle = parsedFrame.Angle
			}
		case "sendMsgToServer":
			playerName := strconv.Itoa(socketId)
			msg := GenericResponse{
				Action: "addToChat",
				Data:   playerName + ": " + parsedFrame.InputId,
			}
			msgByte, err := json.Marshal(msg)
			if err != nil {
				log.Println("error:", err)
			}
			for _, sock := range g.socketList {
				ws.Send(sock, msgByte)
			}
		case "evalServer":

		}
	}
}

func (g *Game) removeSocket(socketId int) {

	g.socketListMutex.Lock()
	defer g.socketListMutex.Unlock()
	log.Println("got or socket lock")

	g.connectedUsers--
	delete(g.socketList, socketId)
	delete(g.playerList, socketId)
}

func (g *Game) addSocket(id int, c *net.Conn) {
	g.socketListMutex.Lock()
	defer g.socketListMutex.Unlock()
	g.connectedUsers++
	g.socketList[id] = c
}
