package handle

import (
	"encoding/base64"
	"encoding/json"
	"github.com/jasonlvhit/gocron"
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
	pingWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
)

func init() {
	log.Println("### Start Ping Scheduler ###")
	gocron.Every(1).Minute().Do(pingActiveDevices)
	gocron.Start()
}

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

var (
	WsConnections     = make(map[string]*websocket.Conn)
	WsAsignConns	  = make(map[string]string)
	//configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")
	//cfg               = config.LoadConfig(*configurationPath)
	WsAllowedOrigin   = "HomeControlApp" //homec.Cfg.WsAllowedOrigin
	IsControlSessionOpen = false;
	WsPresButtFlag = false;
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		//deviceOrigin := r.Header.Get("Origin")
		//if WsAllowedOrigin != deviceOrigin {
		//	log.Println("Origin not allowed:", deviceOrigin)
		//	//log.Println("Origin want:", WsAllowedOrigin)
		//	return false
		//}
		return true
	},
}

func ServeWs(db db.DbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var ws *websocket.Conn
		var err error
		deviceId := r.Header.Get("DeviceId")
		if deviceId =="" {
			deviceId = "br_" + r.Header.Get("Sec-WebSocket-Key")[0:7]
			//
		}

		log.Println("incoming WS request from: ", deviceId)

		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		//defer ws.Close() !!!! Important

		ws.SetCloseHandler(wsClose);

		WsConnections[deviceId] = ws
		if(ws != nil) {
			//ws.SetPingHandler(ping)
			//ws.SetPongHandler(pong)
			log.Println("Create new Ws Connection: succes, device: ", deviceId)
			go wsProcessor(ws, db, deviceId)
		} else {
			log.Println("Ws Connectionfor device: ", deviceId, " not created.")
		}


	}
}

func wsClose(code int, reason string) error  {


	return nil
}

func ping(appData string) error {
	log.Println("[WS] Ping: "+appData)
	return nil
}

func pong(appData string) error {
	log.Println("[WS] Pong: "+appData)
	return nil
}

func wsProcessor(c *websocket.Conn, db db.DbService, dId string) {
	defer c.Close()
	devId := getDevIdByConn(c)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("[WS]:error:  ", err)
			WsConnections[dId] = nil
			break
		}
		log.Printf("[WS]:recive: %s, type: %d", message, mt)
		msg := string(message)

		if strings.Contains(msg, "\"action\":\"connect\"") {
			if !sendMsg(c, "{\"action\":\"connect\",\"success\":true}") {
				break
			} else {
				if strings.Contains(msg, "\"assign\":") {
					assign := msg[30:len(msg)-2]
					WsAsignConns[devId] = assign
					log.Println("#### Assign", assign, " to ", devId)
				}
				log.Println("{action:connect,success:true}", c)
			}
		}
		if strings.Contains(msg, "\"action\":\"resend\"") {
			var rMsg ent.ResendMessage

			err = json.Unmarshal([]byte(msg), &rMsg)
			if err != nil {
				log.Println("Error data unmarshaling: ", err)
			}

			rConn := WsConnections[rMsg.RECIPIENT]
			sender := getDevIdByConn(c)

			message, _ := base64.URLEncoding.DecodeString(rMsg.MESSAGE)
			mm := strings.Replace(string(message),"\"sender\":\"\"", "\"sender\":\""+sender+"\"", 1)
			log.Println("[WS]:resend message: ",  mm, " to: ",  rMsg.RECIPIENT)
			sendMsg(rConn, mm)
		}

		if strings.Contains(msg, "\"action\":\"datasend\"") {

			if strings.Contains(msg, "\"type\":\"koteldata\"") {
				var data ent.KotelData

				wsData := ent.WsSendData{"", "", "", ""}

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

				if !sendMsg(c, "{action:datasend,success:true}") {
					break
				} else {

					err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}
					_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}

					sendDataToWeb(msg, getDevIdByConn(c));

				}

			}

			if strings.Contains(msg, "\"type\":\"roomdata\"") {
				var data ent.SensorsData

				wsData := ent.WsSendData{"", "", "", ""}

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

				data.DATE = time.Now();

				if !sendMsg(c, "{action:datasend,success:true}") {
					break
				} else {
					success, err := db.AddRoomData(data)
					if err != nil || !success {
						log.Println("Error data writing in db: ", err)
					}
				}

				sendDataToWeb(msg, getDevIdByConn(c));

			}
		}
		if strings.Contains(msg, "\"action\":\"pessButton\"") {
			if strings.Contains(msg, "true") {
				WsPresButtFlag = false;
			}

		}
		if strings.Contains(msg, "\"action\":\"sessionStart\"") {
			if strings.Contains(msg, "true") {
				IsControlSessionOpen = true;
			}

		}
		if strings.Contains(msg, "\"action\":\"sessionStop\"") {
			if strings.Contains(msg, "true") {
				IsControlSessionOpen = false;
			}

		}
		if strings.Contains(msg, "\"action\":\"getLastValues\"") {
			if strings.Contains(msg, "\"type\":\"koteldata\"") {
				kd, err := db.GetKotelData()
				log.Println("Get last data from db: ", kd)
				if err != nil {
					log.Println("Error while read data from data base: ", err)
				}

				kData, err := json.Marshal(kd)
				if err != nil {
					log.Println("Error data marshaling: ", err)
				}

				sData := string(kData)
				msg := "{\"action\":\"setLastValues\"," + sData[1:len(sData)]
				log.Println("Last Data is: ", msg)
				if !sendMsg(c, msg) {
					break
				}
			}
			if strings.Contains(msg, "\"type\":\"roomdata\"") {

			}

		}
		if strings.Contains(msg, "\"action\":\"getMeshValues\"") {
			if strings.Contains(msg, "\"type\":\"koteldata\"") {

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
	devId := getDevIdByConn(c)
	log.Println("[WS]:send to " + devId + " success: "+m)
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


func pingActiveDevices() {
	log.Println("########## Ping Devices ############")
	for devId, c := range WsConnections {
		if c != nil {
			if !sendMsg(c, "{\"action\":\"ping\"}") {
				log.Println("# Send ping to " + devId + " failed.  #")
				deleteWsConn(devId)
			} else {
				log.Println("# Send ping to " + devId + " success. #" )
			}
		} else {
			log.Println("# Device " + devId + " is die.        #" )
			deleteWsConn(devId)
		}

	}
	log.Println("####################################")
}

func deleteWsConn(dId string) {
	WsConnections[dId] = nil
	delete(WsConnections, dId)
	delete(WsAsignConns, dId)
}

func sendDataToWeb(msg string, sender string) {
	for devId, c := range WsConnections {
		if c==nil {
			deleteWsConn(devId);
			break;
		}
		assDev := getDevIdByConn(c)
		assRcp := WsAsignConns[assDev]
		if strings.Contains(devId, "br_") && sender==assRcp {
			if !sendMsg(c, msg) {
				log.Println("# Send data to " + devId + " failed.  #")
				//deleteWsConn(devId)
			} else {
				log.Println("# Send data to " + devId + " success. #" )
			}
		}
	}
}