# Small Chat Server

Ein einfacher Chat-Server, der Nachrichten über WebSockets liest und an Clients sendet.

## Funktionen

- Echtzeit-Nachrichtenübertragung über WebSockets
- Unterstützung für mehrere Chat-Räume mit eindeutigen Chat-IDs
- Speicherung von Nachrichten in MongoDB
- Emoji-Reaktionen auf Nachrichten
- Einfache Integration in Webseiten

## Technologien

- **Backend**: Go 1.21
- **WebSockets**: Gorilla WebSocket
- **Datenbank**: MongoDB
- **Konfiguration**: godotenv für Umgebungsvariablen

## Installation

### Voraussetzungen

- Go 1.21 oder höher
- MongoDB (lokal oder über Docker)

### Schritte

1. Repository klonen:
   ```bash
   git clone https://github.com/yourusername/small-chat-server.git
   cd small-chat-server
   ```

2. Abhängigkeiten installieren:
   ```bash
   go mod download
   ```

3. `.env` Datei erstellen:
   ```
   MONGO_USER=admin
   MONGO_PASS=admin
   MONGO_HOST=localhost:27017
   ```

4. MongoDB starten (mit Docker):
   ```bash
   docker build -t chat-mongodb .
   docker run -d -p 27017:27017 --name chat-db chat-mongodb
   ```

5. Server starten:
   ```bash
   go run main.go
   ```

## Verwendung

### Server starten

```bash
go run main.go
```

Der Server startet auf Port 8080 und ist bereit, WebSocket-Verbindungen anzunehmen.

### Test-Client

Ein einfacher Test-Client ist unter `temp/index.html` verfügbar. Öffnen Sie diese Datei in einem Browser, um die Chat-Funktionalität zu testen.

### WebSocket-Integration

Um den Chat-Server in Ihre eigene Anwendung zu integrieren:

```javascript
let createWebSocket = (chatId) => {
  socket = new WebSocket(`ws://localhost:8080/chat?chat_id=${chatId}`);

  socket.onopen = () => {
    console.log('Socket opened');
  };

  socket.onmessage = (event) => {
    console.log(event.data);
    // Nachricht verarbeiten (JSON-Format)
  };

  socket.onclose = () => {
    console.log('Socket closed');
  };

  socket.onerror = (error) => {
    console.log(error);
  };
};

// Nachricht senden
function sendMessage(sender, content) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({
      sender: sender,
      content: content
    });
    socket.send(message);
  }
}

// Reaktion senden
function sendReaction(messageId, emoji, sender) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const reaction = JSON.stringify({
      type: "reaction",
      messageId: messageId,
      emoji: emoji,
      sender: sender
    });
    socket.send(reaction);
  }
}
```

## API-Endpunkte

- `ws://localhost:8080/chat?chat_id=<CHAT_ID>` - WebSocket-Endpunkt für Chat-Verbindungen
- `http://localhost:8080/health` - Gesundheitscheck für die Datenbankverbindung

## Nachrichtenformat

### Reguläre Nachricht
```json
{
  "sender": "username",
  "content": "Nachrichtentext",
  "type": "message"
}
```

### Reaktion
```json
{
  "type": "reaction",
  "messageId": "sender-content",
  "emoji": "👍",
  "sender": "username"
}
```
