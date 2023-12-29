# small-chat-server


This is small Chat Server which reads and send messages via websocket
to a Client which can be website.


### Integration
````js
let createWebSocket = () => {
  socket = new WebSocket('ws://localhost:8080/chat');

  socket.onopen = () => {
    console.log('Socket opened');
  };
  socket.onmessage = (event) => {
    console.log(event.data);
  };

  socket.onclose = () => {
    console.log('Socket closed');
  };
  socket.onerror = (error) => {
    console.log(error);
  };
};
````

