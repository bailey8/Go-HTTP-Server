package httpServer

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	db "cse312.app/database"
	"cse312.app/game"
	util "cse312.app/utility"
	"cse312.app/values"
)

func Login(c net.Conn, req *Request) {
	result := db.VerifyCredentials(string(req.PostData["username"]), string(req.PostData["password"]))
	switch result {
	case true:
		token := util.GenerateToken()
		db.StoreToken(string(req.PostData["username"]), token)
		//util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"], values.Headers["content-text"],
		//fmt.Sprintf("Set-Cookie: id=%s; HttpOnly\r\n", token)}, nil)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"], fmt.Sprintf("Set-Cookie: id=%s; HttpOnly\r\n", token)}, []byte("true"))
	case false:
		//util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-index"], values.Headers["content-text"]}, nil)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("false"))
	}
}

func Register(c net.Conn, req *Request) {
	result := db.RegisterUser(string(req.PostData["username"]), string(req.PostData["password"]))
	switch result {
	case true:
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("<h1> Registered Successfully </h1>"))
	case false:
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("<h1> Registration Failed </h1>"))
	}
}

//this creates the lobby
//the client will then have to use join lobby which uses a query string
func CreateLobby(c net.Conn, req *Request) {
	if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
		util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
		return
	}

	lobbyId := string(req.PostData["lobbyId"])
	lobby := game.GetLobby(lobbyId)
	if lobby == nil {
		lobby = game.MakeLobby(&c, lobbyId)
		template, _ := ioutil.ReadFile("index.html")
		template = bytes.Replace(template, []byte("{{response1}}"), []byte(fmt.Sprintf("<h1> Success! Your lobby id is %s </h1>", req.PostData["lobbyId"])), 1)
		template = bytes.Replace(template, []byte("{{response2}}"), []byte(""), 1)
		//util.SendResponse(c, []string{values.Headers["200"], values.Headers["redirect-index"], values.Headers["content-html"]}, template)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte(fmt.Sprintf("<h1> Success! Your lobby id is %s </h1>", lobbyId)))
	} else {
		template, _ := ioutil.ReadFile("index.html")
		template = bytes.Replace(template, []byte("{{response1}}"), []byte("<h1> Error! Lobby id is already in use</h1>"), 1)
		template = bytes.Replace(template, []byte("{{response2}}"), []byte(""), 1)
		//util.SendResponse(c, []string{values.Headers["200"], values.Headers["redirect-index"], values.Headers["content-html"]}, template)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("<h1> Error! Lobby id is already in use</h1>"))
	}
	//log.Panic(lobbyId, req.PostData)
}

func UploadProfilePic(c net.Conn, req *Request) {
	if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
		util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
		return
	}

	fileName := "images/" + fmt.Sprintf("%x", sha1.Sum(req.PostData["upload"])) + ".jpg" //TODO

	_, username := db.IsValidToken(req.Cookies["id"])
	db.StoreProfilePath(username, fileName)

	err := ioutil.WriteFile(fileName, req.PostData["upload"], 0644)
	if err != nil {
		log.Panic(err)
	}
	
	util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-index"], values.Headers["content-html"]}, nil)
}
