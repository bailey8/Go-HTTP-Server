// Establish a WebSocket connection with the server
const activeUsers = new WebSocket('ws://' + window.location.host + '/websocket/active_users');
activeUsers.onmessage = addMessage;

// Allow users to send messages by pressing enter instead of clicking the Send button
document.addEventListener("keypress", function (event) {
    if (event.code === "Enter") {
        sendMessage();
    }
 });
 
 // Read the name/comment the user is sending to chat and send it to the server over the WebSocket as a JSON string
 // Called whenever the user clicks the Send button or pressed enter
 function sendMessage() {
    msg = document.getElementById("chat-comment").value;
    const chatBox = document.getElementById("chat-comment");
    const comment = chatBox.value;
    chatBox.value = "";
    chatBox.focus();
    if(comment === null || comment.trim() === "") {
        return
    }

    msg = msg.trim();
    activeUsers.send(JSON.stringify({'Action': 'broadcastMsg', 'ChatMsg': msg, 'Receiver': ''}));
    return

    //no unicast messages from chatbox
    /*index = msg.indexOf('@');
    
    receiver = '';

    if (index === -1) {
        activeUsers.send(JSON.stringify({'Action': 'broadcastMsg', 'ChatMsg': msg, 'Receiver': ''}));
        return
    }
    index2 = msg.indexOf(' ');
    //TODO error checking
    receiver = (index2 === -1) ?  msg.substr(1,) :  msg.substr(1,index2).trim()
    
    activeUsers.send(JSON.stringify({'Action': 'unicastMsg', 'ChatMsg': msg.substr(1,), 'Receiver': receiver}));
    */
   
 }
 
 // Called when the server sends a new message over the WebSocket and renders that message so the user can read it
 function addMessage(message) {
    const chatMessage = JSON.parse(message.data);
    switch (chatMessage.Action) {
        case 'displayUsers':
            let users = document.getElementById('activeUsers');
     
            prev = users.innerHTML
            users.innerHTML = ''
            for(var i = 0 ; i < chatMessage.Users.length; i++) {
                /*var	request	=  new	XMLHttpRequest();	
                request.username = chatMessage.Users[i].ProfilePic
                request.onreadystatechange = function() {	
	           	 if	(this.readyState === 4 && this.status === 200){	
	           		console.log(this.response);	
	           	 }	
                };	
                request.open("GET","/image?username="+chatMessage.Users[i].ProfilePic);	
                let	data = {'username':	"Jesse", 'message':	"Welcome"}	
                request.send(JSON.stringify(data));*/
                users.innerHTML += `<div style="cursor:pointer;" onClick="sendDM(this)" id="`+chatMessage.Users[i].Username+`"> <img src="`+chatMessage.Users[i].ProfilePic + `"class="img-circle rounded float-left" width="50px" height="50px"  style="float: left;"/> <h6>`  + chatMessage.Users[i].Username + '</h6> <br>'
            }
            break;
        case 'displayBroadcast':
           msg = chatMessage.ChatMsg;
           chat.innerHTML += '<div style="float: left;"> ' + msg + "</div> <br>";
           break;
        case 'displayUnicast':
           msg = chatMessage.ChatMsg;
           //chat.innerHTML += '<div style="float: left; background-color: green; opacity: 0.3;"> ' + msg + "</div> <br>";
           if (chatMessage.Alert === false ) {
            return
           }
           if (chatMessage.Alert === true ) {
               display = chatMessage.Sender + ' sent you a message:\n'+ msg
               alert(display);
           } 
           user = chatMessage.Sender;
           msg = window.prompt("Send message to ", user);
            if (msg === null || msg.trim() === "") {
                return
            }
            activeUsers.send(JSON.stringify({'Action': 'unicastMsg', 'ChatMsg': msg, 'Receiver': user}));
            break;
     }
     console.log('made it')
 }

 function sendDM(div) {
    user = div.id;
    msg = window.prompt("Send message to ", user);
    if (msg === null || msg.trim() === "") {
        return
    }
    activeUsers.send(JSON.stringify({'Action': 'unicastMsg', 'ChatMsg': msg, 'Receiver': user}));
  }

 function createLobby() {
    const form = new FormData(document.getElementById('createLobby-form'));
    form.get('lobbyId')
   fetch('/createLobby', {
       method: 'POST',
       body: form
   }).then(function(response) { //https://developer.mozilla.org/en-US/docs/Web/API/Body/text
    return response.text().then(function(text) {
        console.log("text", text);
        elem = document.getElementById('status1');
        elem.innerHTML = text;
    });
   });
}

function uploadPic() {
    const form = new FormData(document.getElementById('createLobby-form'));
    form.get('lobbyId')
   fetch('/createLobby', {
       method: 'POST',
       body: form
   }).then(function(response) { //https://developer.mozilla.org/en-US/docs/Web/API/Body/text
    return response.text().then(function(text) {
        console.log("text", text);
        elem = document.getElementById('status1');
        elem.innerHTML = text;
    });
   });
}
 
 
 function joinGame() {
     id =  document.getElementById('joinLobbyId').value;
     window.location.href = '/game?lobbyId='+id
 }
 
 function changeBackground() {
     r = Math.floor(Math.random() * 122) + 100; //keeps colors light 
     g = Math.floor(Math.random() * 122) + 100; //keeps colors light 
     b = Math.floor(Math.random() * 122) + 100; //keeps colors light 
     document.body.style.backgroundColor = 'rgb(' + r + ',' + g + ',' + b + ')';
 }
 
 function sendImage() {
     const chatName = document.getElementById("chat-name").value;
     const file = document.getElementById("chat-image").files[0];
     b = new Blob([file])
     socket.send(b)
 }


 setInterval(function() {
    activeUsers.send(JSON.stringify({'Action': 'displayUsers'}) );
  }, 5000 ); 