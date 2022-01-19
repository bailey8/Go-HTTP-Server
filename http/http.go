package httpServer

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	db "cse312.app/database"
	util "cse312.app/utility"
	"cse312.app/values"
)

type Request struct {
	Headers      map[string]string
	Payload      []byte
	ReqType      string
	Path         string
	Cookies      map[string]string
	PostData     map[string][]byte
	QueryStrings map[string][]string
}

func HandleConnection(c net.Conn) {
	defer c.Close()
	data := make([]byte, 2048) //NOTE, buffer must be large enough to read full HTTP headers --> Fixed
	n, err := c.Read(data)
	//removes excess bytes - i.e. x00
	data = data[:n]
	if err != nil {
		fmt.Println("Socket closed")
		c.Close()
		return
	}

	req := parseRequest(c, data)

	switch req.ReqType {
	case "GET":
		handleGET(c, req)
	case "POST":
		handlePOST(c, req)
	default:
		fmt.Println("Unknown HTTP request")
	}
}

func parseRequest(c net.Conn, data []byte) *Request {
	req := Request{Headers: make(map[string]string), Cookies: make(map[string]string),
		PostData: make(map[string][]byte)}

	//TODO this might be an error --> data might not be appended to data slice
	headerEnds := util.GetHeaderLoc(data, c)
	if headerEnds == -1 {
		//TODO replace panic with response
		log.Panic("no body")
	}
	headers := strings.Split(string(data[:headerEnds]), "\r\n")

	//assign payload
	if headerEnds >= len(data) {
		req.Payload = nil
	} else {
		req.Payload = data[headerEnds+len([]byte("\r\n\r\n")):]
	}

	//extract GET or POST type
	req.ReqType = strings.Split(headers[0], " ")[0]

	//extract path
	req.Path = strings.Split(headers[0], " ")[1]

	//check for query string and extract it
	if strings.Contains(req.Path, "?") == true {
		req.QueryStrings = util.ParseQuery(strings.Split(req.Path, "?")[1])

		//trim path for switch statement
		index := strings.Index(req.Path, "?")
		req.Path = req.Path[:index]
	}

	//TODO error --> no content-type when user resends form data when they refresh page (on Firefox) --> RESOLVED
	log.Println(string(data))
	//starts at 1 to skip GET/POST header
	for i := 1; i < len(headers); i++ {
		split := strings.Split(headers[i], ": ")
		key, value := split[0], split[1]
		req.Headers[key] = value
	}

	//get Cookie info
	if req.Headers["Cookie"] != "" {
		req.Cookies = util.ParseCookie(req.Headers["Cookie"])
	}

	//get POST data
	if req.ReqType == "POST" {
		//we need to first extract the web-kit boundary and then read ALL of the data before processing payload
		expectedBytes, _ := strconv.Atoi(req.Headers["Content-Length"])
		boundary := strings.Split(req.Headers["Content-Type"], "boundary=")[1]

		//read more data if the full payload wasn't in POST request
		currentBytes := len(req.Payload)
		if expectedBytes > len(req.Payload) {
			for expectedBytes > currentBytes {
				buff := make([]byte, expectedBytes-currentBytes)
				n, _ := c.Read(buff)
				buff = buff[:n]
				req.Payload = append(req.Payload, buff...)
				//removes excess bytes - i.e. x00
				currentBytes += n
			}
		}
		log.Println("parsing data")
		req.PostData = util.ParseMultiForm(req.Payload, boundary)
	}

	return &req
}

func handleGET(c net.Conn, req *Request) {
	switch req.Path {
	case "/":
		Home(c, req)
	case "/home":
		Home(c, req)
	case "/landing":
		Landing(c, req)
	case "/game":
		Game(c, req)
	case "/ws_game":
		WS_Game(c, req)
	case "/websocket/active_users":
		WS_ActiveUsers(c, req)
	case "/current_users":
		ActiveUsers(c, req)
	case "/get_profile_picture_path":
		GetProfilePath(c, req)
		//////////////////////////////////////////////////////////////////////////////////////////
	case "/api/current_user":
		//send application/json data back with user info
		CurrentUser(c, req)
		//////////////////////////////////////////////////////////////////////////////////////////
	case "/api/cookie":
		//////////////////////////////////////////////////////////////////////////////////////////
	default:
		//trim '/' from file name
		req.Path = strings.TrimLeft(req.Path, "/")
		log.Println(req.Path)
		if _, err := os.Stat(req.Path); err == nil && values.ValidFiles[req.Path] == true { //handles specific requests for files
			util.SendFile(c, req.Path)
			return
		}

		//make sure user is logged-in before requesting user-specific content
		if result, _ := db.IsValidToken(req.Cookies["id"]); result == false {
			util.SendResponse(c, []string{values.Headers["301"], values.Headers["redirect-home"]}, nil)
			return
		}

		//check for profile images
		log.Println("HERE", req.Path)
		log.Println("HERE234", req.Path)
		_, err := os.Stat(req.Path)
		log.Println(err, os.IsNotExist(err))
		if /*_, err := os.Stat(req.Path); /*os.IsNotExist(err) &&*/ strings.Contains(req.Path, "images/") { //TODO
			req.Path = strings.ReplaceAll(req.Path, "..", "nope") //TODO replace ~ too
			req.Path = strings.ReplaceAll(req.Path, "~", "nope")
			util.SendFile(c, req.Path)
			return
		}
		//robust method to check existence of file for React components
		if _, err := os.Stat(req.Path); os.IsNotExist(err) && strings.Contains(req.Path, "/client/build") {
			req.Path = strings.ReplaceAll(req.Path, "..", ".") //TODO replace ~ too
			util.SendFile(c, req.Path)
		} else {
			util.SendFile(c, "client/build/index.html")
		}
	}
}

func handlePOST(c net.Conn, req *Request) {

	switch req.Path {
	case "/login":
		Login(c, req)
	case "/register":
		Register(c, req)
	case "/createLobby":
		CreateLobby(c, req)
	case "/uploadProfilePic":
		UploadProfilePic(c, req)
	}

}
