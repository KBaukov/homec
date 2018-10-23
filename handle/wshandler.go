package handle

import (
	"encoding/json"
	"homec/db"
	"homec/ent"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second 		// Time allowed to write a message to the peer.
	maxMessageSize = 8192  				// Maximum message size allowed from peer.
	pongWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
)

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

var (
	WsConnections     = make(map[string]*websocket.Conn)
	//configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")
	//cfg               = config.LoadConfig(*configurationPath)
	WsAllowedOrigin   = "HomeControlApp" //homec.Cfg.WsAllowedOrigin
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		deviceOrigin := r.Header.Get("Origin")
		if WsAllowedOrigin != deviceOrigin {
			log.Println("Origin not allowed:", deviceOrigin)
			log.Println("Origin want:", WsAllowedOrigin)
			return false
		}
		return true
	},
}

func ServeWs(db db.DbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var ws *websocket.Conn
		var err error
		deviceId := r.Header.Get("DeviceId")

		log.Println("incoming request from: ", deviceId)

		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		//defer ws.Close() !!!! Important

		WsConnections[deviceId] = ws

		log.Println("Ws Connection: ", ws)
		go wsProcessor(ws, db)

	}
}

func wsProcessor(c *websocket.Conn, db db.DbService) {
	defer c.Close()
	devId := getDevIdByConn(c)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("[WS]:recv:", err)
			break
		}
		log.Printf("[WS]:recv: %s, type: %d", message, mt)
		msg := string(message)

		if strings.Contains(msg, "{action:connect") {
			if !sendMsg(c, "{action:connect,success:true}") {
				break
			}
		}

		if strings.Contains(msg, "{\"action:\":\"datasend\"") {

			if strings.Contains(msg, "\"type\":\"koteldata\"") {
				var data ent.KotelData

				wsData := ent.WsSendData{"", "", ""}

				err = json.Unmarshal([]byte(msg), &wsData)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				dd, err := json.Marshal(wsData.DATA)
				if err != nil {
					log.Println("Error data marshaling: ", err)
				}

				err = json.Unmarshal(dd, &data)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				log.Println("get kotel data:", data)

				if !sendMsg(c, "{action:datasend,success:true}") {
					log.Println("Send to " + devId + ": failed")
					break
				} else {
					log.Println("Send to " + devId + ": success")
					err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}
				}

			}
		}

		if strings.Contains(msg, "Important") {

			if !sendMsg(c, "Ok. I'm...") {
				break
			}
		}

	}

}

func sendMsg(c *websocket.Conn, m string) bool {
	err := c.WriteMessage(1, []byte(m))
	if err != nil {
		log.Println("[WS]:send:", err)
		return false
	}
	return true
}

func getDevIdByConn(c *websocket.Conn) string {
	for k, v := range WsConnections {
		if v == c {
			return k
		}
	}

	return ""
}


