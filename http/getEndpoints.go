package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	db "cse312.app/database"
	"cse312.app/game"
	util "cse312.app/utility"
	"cse312.app/values"
	websocket "cse312.app/websocket"
)

func Home(c net.Conn, req *Request) {

	result, username := db.IsValidToken(req.Cookies["id"])
	switch result {
	case true:
		template, _ := ioutil.ReadFile("index.html")
		template = bytes.Replace(template, []byte("{{pic}}"), []byte(db.GetProfilePath(username)), 1)
		template = bytes.Replace(template, []byte("{{response1}}"), []byte(""), 1)
		template = bytes.Replace(template, []byte("{{response2}}"), []byte(""), 1)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-html"]}, template)
	case false:
		template, _ := ioutil.ReadFile("login.html")
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-html"]}, template)
	}
}

func Landing(c net.Conn, req *Request) {
	//make sure client is logged-in before processing request
	if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
		util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
		return
	}
	template, _ := ioutil.ReadFile("landing.html")
	util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-html"]}, template)
}

func Game(c net.Conn, req *Request) {
	if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
		util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
		return
	}
	template, _ := ioutil.ReadFile("game.html")
	util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-html"]}, template)
}

func JoinLobby(c net.Conn, req *Request) {
	if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
		util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
		return
	}
	//TODO use ajax later
	lobby := game.GetLobby(req.QueryStrings["lobbyId"][0])
	template, _ := ioutil.ReadFile("index.html")
	template = bytes.Replace(template, []byte("{{response1}}"), []byte(""), 1)

	if lobby == nil {
		template = bytes.Replace(template, []byte("{{response2}}"), []byte("<h1> Error! Lobby does not exists; please make the lobby first</h1>"), 1)
	} else if lobby.GetNumPlayers() >= game.MAX_LOBBY_SIZE {
		template = bytes.Replace(template, []byte("{{response2}}"), []byte("<h1> Error! Lobby is full; please wait or make a new lobby</h1>"), 1)
	} else {

	}
	util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-html"]}, template)
}

func WS_Game(c net.Conn, req *Request) {
	//here are the rules:
	//if lobby doesn't exist then make new lobby and join it
	//if lobby exists, make sure numUsers in lobby < MAX_USERS -->MAX_USERS = 2
	//	if lobby is full, then send reject request
	ok, username := db.IsValidToken(req.Cookies["id"])
	if ok == false {
		return
	}
	lobby := game.GetLobby(req.QueryStrings["lobbyId"][0])
	if lobby == nil {
		return
		//lobby = game.MakeLobby(&c, req.QueryStrings["lobbyId"][0])
	}
	err := lobby.AddPlayer(&c)
	if err != nil {
		//lobby is full
		return
	}

	key := req.Headers["Sec-WebSocket-Key"]
	if key == "" {
		log.Panic("didn't find key")
	}
	lobby.GameInstance.PlayGame(c, key, username)
}

//returns json data of all active users
func ActiveUsers(c net.Conn, req *Request) {
	//return all users in values.UpgradedConn
}

type Frame struct { //different from frame used by the game
	Action   string
	ChatMsg  string
	Sender   string
	Receiver string
	Alert    bool
}

func WS_ActiveUsers(c net.Conn, req *Request) {

	//parse websocket request
	ok, username := db.IsValidToken(req.Cookies["id"])
	if ok == false {
		return
	}
	key := req.Headers["Sec-WebSocket-Key"]
	if key == "" {
		log.Panic("didn't find key")
	}
	ws := websocket.UpgradeConn(c, key, username)
	defer ws.Close(username)

	for {
		frame := <-ws.GetChan()
		if frame == nil {
			return
		}

		var parsedFrame Frame
		err := json.Unmarshal(frame.Payload, &parsedFrame)
		_ = err
		log.Println(parsedFrame.Action)
		switch parsedFrame.Action {
		case "displayUsers":
			//get users
			users := websocket.GetActiveUsers()
			msg := websocket.ActiveUsersJson{
				Action: "displayUsers",
				Users:  users,
			}
			msgByte, err := json.Marshal(msg)
			if err != nil {
				log.Println("error:", err)
			}
			ws.Send(&c, msgByte)
		case "broadcastMsg":
			log.Println(parsedFrame.Receiver, parsedFrame.ChatMsg)

			var chat string
			chat = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
				fmt.Sprintf("%s: %s\t", username, parsedFrame.ChatMsg),
				"&", "&amp"),
				"<", "&lt"),
				">", "&gt")
			log.Println(chat, username)
			//add msg to chatHistory
			values.ChatHistoryMutex.Lock()
			values.ChatHistory = append(values.ChatHistory, []byte(chat)...)
			values.ChatHistoryMutex.Unlock()

			msg := Frame{
				Action:   "displayBroadcast",
				ChatMsg:  chat,
				Receiver: "",
				Alert:    false,
			}
			msgByte, err := json.Marshal(msg)
			if err != nil {
				log.Println("error:", err)
			}

			for _, socket := range websocket.GetActiveSockets() {
				ws.Send(socket, msgByte)
			}
		case "unicastMsg":
			chat := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
				fmt.Sprintf("%s", parsedFrame.ChatMsg),
				"&", "&amp"),
				"<", "&lt"),
				">", "&gt")
			msg := Frame{
				Action:   "displayUnicast",
				ChatMsg:  chat,
				Receiver: parsedFrame.Receiver,
				Alert:    false,
				Sender:   username,
			}
			log.Println("Receiver = ", parsedFrame.Receiver)
			msgByte, err := json.Marshal(msg)
			if err != nil {
				log.Println("error:", err)
			}
			log.Println(parsedFrame.Sender, parsedFrame.Receiver, parsedFrame.Alert)
			//ws.Send(&c, msgByte)
			//get ALL sockets associated with sender
			log.Println("sender = ", username)
			sockets := websocket.GetUsersSockets(username)
			for _, socket := range sockets {
				ws.Send(socket, msgByte)
			}

			/*
				msg = Frame{
					Action:   "displayUnicast",
					ChatMsg:  chat,
					Receiver: parsedFrame.Receiver,
					Alert:    true,
					Sender:   username,
				}
			*/
			msg.Alert = true
			log.Println("Receiver = ", parsedFrame.Receiver)
			msgByte, err = json.Marshal(msg)
			if err != nil {
				log.Println("error:", err)
			}
			//get ALL sockets associated with receiver
			log.Println("Receiver = ", parsedFrame.Receiver)
			sockets = websocket.GetUsersSockets(parsedFrame.Receiver)
			for _, socket := range sockets {
				ws.Send(socket, msgByte)
			}
		}
	}
}

func GetProfilePath(c net.Conn, req *Request) {
	//TODO do error checking
	if len(req.QueryStrings["username"]) == 0 {
		log.Panic()
	}
	//not the most optimal solution because we can't cache
	path := db.GetProfilePath(req.QueryStrings["username"][0])
	util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte(path))
}
