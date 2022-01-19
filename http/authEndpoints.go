package httpServer

import (
	"log"
	"net"

	db "cse312.app/database"
	util "cse312.app/utility"
	"cse312.app/values"
)

/*
	These functions are supposed to mimic the passport api
*/

func CurrentUser(c net.Conn, req *Request) {
	//debug
	log.Println("here")
	util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("Invalid Token. I pity the fool who queries the site w/out a token."))
	return //temporary
	//first determine if token from user is valid to obtain the username
	//if it is, then we query the database again with username to get rest of the data
	//if token is invalid, send blank response
	result, username := db.IsValidToken(req.Cookies["id"])
	switch result {
	case true:
		userDataJSON := db.GetUserInfo(username)
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-json"]}, userDataJSON)
	case false:
		util.SendResponse(c, []string{values.Headers["200"], values.Headers["content-text"]}, []byte("Invalid Token. I pity the fool who queries the site w/out a token."))
	}
}
