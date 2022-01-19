package game

import (
	"errors"
	"net"
	"sync"
)

var MAX_LOBBY_SIZE = 2
var Lobbies = make(map[string]*Lobby)
var LobbiesMutex = &sync.Mutex{}

type Lobby struct {
	GameInstance        *Game
	ConnectedUsers      map[*net.Conn]bool //TODO - there may be a race condition here
	ConnectedUsersMutex *sync.Mutex
}

func MakeLobby(c *net.Conn, lobbyId string) *Lobby {
	LobbiesMutex.Lock()
	defer LobbiesMutex.Unlock()

	if Lobbies[lobbyId] != nil { //lobby already exists
		return nil
	}
	l := &Lobby{GameInstance: NewGame(), ConnectedUsers: make(map[*net.Conn]bool, 0), ConnectedUsersMutex: &sync.Mutex{}}
	Lobbies[lobbyId] = l

	//need for cspGame
	//go l.GameInstance.GameRunner()
	return l

}

func GetLobby(id string) *Lobby {
	LobbiesMutex.Lock()
	defer LobbiesMutex.Unlock()
	return Lobbies[id]
}

func (l *Lobby) GetNumPlayers() int {
	return l.GameInstance.connectedUsers
}

func (l *Lobby) AddPlayer(c *net.Conn) error {
	if l.GameInstance.connectedUsers < MAX_LOBBY_SIZE {
		//l.ConnectedUsers = append(l.ConnectedUsers, c)
		l.ConnectedUsersMutex.Lock()
		defer l.ConnectedUsersMutex.Unlock()
		l.ConnectedUsers[c] = true

		return nil
	}
	return errors.New("Lobby Full")
}

func (l *Lobby) RemovePlayer(c *net.Conn) error {
	l.ConnectedUsersMutex.Lock()
	defer l.ConnectedUsersMutex.Unlock()

	//l.ConnectedUsers = append(l.ConnectedUsers, c)
	delete(l.ConnectedUsers, c)
	return nil
}
