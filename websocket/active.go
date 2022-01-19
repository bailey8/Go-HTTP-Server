package websocket

import (
	"net"
	"sync"

	db "cse312.app/database"
)

//a user is logged in if they have at least one active ws connec
var ActiveUsers = make(map[string]int)
var ActiveSockets = make(map[*net.Conn]string)
var ActiveUsersMutex = &sync.Mutex{}

type ActiveUsersJson struct {
	Action string
	Users  []UserData
}

type UserData struct {
	Username   string
	ProfilePic string
}

func addUser(user string, c *net.Conn) {
	ActiveUsersMutex.Lock()
	defer ActiveUsersMutex.Unlock()

	ActiveSockets[c] = user
	ActiveUsers[user]++
}

func removeUser(user string, c *net.Conn) {
	ActiveUsersMutex.Lock()
	defer ActiveUsersMutex.Unlock()

	delete(ActiveSockets, c)
	ActiveUsers[user]--
	if ActiveUsers[user] <= 0 {
		delete(ActiveUsers, user)
	}
}

//returns all active users
func GetActiveUsers() []UserData {
	ActiveUsersMutex.Lock()
	defer ActiveUsersMutex.Unlock()

	res := make([]UserData, 0)
	for user := range ActiveUsers {
		res = append(res, UserData{Username: user, ProfilePic: db.GetProfilePath(user)})
	}
	return res
}

func GetActiveSockets() []*net.Conn {
	ActiveUsersMutex.Lock()
	defer ActiveUsersMutex.Unlock()

	res := make([]*net.Conn, 0)
	for user := range ActiveSockets {
		res = append(res, user)
	}
	return res
}

func GetUsersSockets(username string) []*net.Conn {
	ActiveUsersMutex.Lock()
	defer ActiveUsersMutex.Unlock()

	res := make([]*net.Conn, 0)
	for sock, user := range ActiveSockets {
		if user == username {
			res = append(res, sock)
		}
	}
	return res
}
