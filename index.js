// Establish a WebSocket connection with the server
const socket = new WebSocket('ws://' + window.location.host + '/websocket/active_users');

// Call the addMessage function whenever data is received from the server over the WebSocket
//socket.onmessage = addMessage;

// Allow users to send messages by pressing enter instead of clicking the Send button
document.addEventListener("keypress", function (event) {
    if (event.code === "Enter") {
        sendMessage();
    }
 });
 
 // Read the name/comment the user is sending to chat and send it to the server over the WebSocket as a JSON string
 // Called whenever the user clicks the Send button or pressed enter
 function sendMessage() {
    const chatName = document.getElementById("chat-name").value;
    const chatBox = document.getElementById("chat-comment");
    const comment = chatBox.value;
    chatBox.value = "";
    chatBox.focus();
    if(comment !== "") {
        socket.send(JSON.stringify({'username': chatName, 'comment': comment}));
    }
 }
 
 // Called when the server sends a new message over the WebSocket and renders that message so the user can read it
 function addMessage(message) {
     chatMessage = ''
     try {
     chatMessage = JSON.parse(message.data);
     let chat = document.getElementById('chat');
     chat.innerHTML += "<b>" + chatMessage['username'] + "</b>: " + chatMessage["comment"] + "<br/>";
    } catch(e) {
     chatMessage = message.data;
     let chat = document.getElementById('chat');
     
     //console.log(this.response);	
     chat.innerHTML += "<b>" + message.data;
 
    }
   
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