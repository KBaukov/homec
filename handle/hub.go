package handle

import (
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

type Conn struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte

	deviceId string
}

func (c *Conn) GetDeviceId() string {
	return c.deviceId
}
// write writes a message with the given message type and payload.
func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

type Hub struct {
	// Registered connections.
	connections map[*Conn]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *Conn

	// Unregister requests from connections.
	unregister chan *Conn
}

var hub = Hub{
	broadcast:   make(chan []byte),
	register:    make(chan *Conn),
	unregister:  make(chan *Conn),
	connections: make(map[*Conn]bool),
}

func (h *Hub) run() {
	log.Println("### Connection Hub is started ###")
	for {
		select {
		case conn := <-h.register:
			h.connections[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.connections[conn]; ok {
				log.Println("### Unregistr: "+conn.deviceId)
				unAssign(conn.deviceId)
				delete(h.connections, conn)
				close(conn.send)
			}
		case message := <-h.broadcast:
			for conn := range h.connections {
				select {
					case conn.send <- message:
					default:
						close(conn.send)
						delete(hub.connections, conn)
				}
			}
		}
	}
}

func (h *Hub) getConnByDevId(devId string) *Conn {
	for conn := range h.connections {
		if conn.deviceId == devId {
			return conn
		}
	}
	return nil
}

func (h *Hub) sendDataToWeb(msg string, sender string) {
	for conn := range h.connections {
		assDev := conn.deviceId
		assRcp := WsAsignConns[assDev]
		if strings.Contains(assDev, brPref) && sender==assRcp {
			if !sendMsg(conn, msg) {
				log.Println("# Send data to " + assDev + " failed.  #")
			} else {
				log.Println("# Send data to " + assDev + " success. #" )
			}
		}
	}
}