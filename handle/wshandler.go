package handle

import (
	//"bytes"
	"encoding/json"
	"github.com/KBaukov/homec/db"
	"github.com/KBaukov/homec/ent"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second 		// Time allowed to write a message to the peer.
	maxMessageSize = 1024  				// Maximum message size allowed from peer.
	pingWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
)

func init() {

	//log.Println("### Ping Scheduler is started ###")
	//gocron.Every(1).Seconds().Do(pingActiveDevices)
	//gocron.Start()
	go hub.run()
}

//func internalError(ws *websocket.Conn, msg string, err error) {
//	log.Println(msg, err)
//	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
//}

var (
	//WsConnections     = make(map[string]*websocket.Conn)
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
		devId := r.Header.Get("DeviceId")
		if devId =="" {
			devId = "user_" + r.Header.Get("Sec-WebSocket-Key")
		}

		log.Println("incoming WS request from: ", devId)

		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		//defer ws.Close() !!!! Important

		conn := &Conn{send: make(chan []byte, 256), ws: ws, deviceId:devId}
		hub.register <- conn
		go conn.writePump(db)
		conn.readPump(db)

		//WsConnections[deviceId] = ws
		//if(ws != nil) {
		//	//ws.SetPingHandler(ping)
		//	//ws.SetPongHandler(pong)
		//	log.Println("Create new Ws Connection: succes, device: ", deviceId)
		//	go wsProcessor(ws, db, deviceId)
		//} else {
		//	log.Println("Ws Connectionfor device: ", deviceId, " not created.")
		//}


	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Conn) readPump(db db.DbService) {
	defer func() {
		hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	devId := c.deviceId

	for {
		mt, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("[WS]:recive: %s, type: %d", message, mt)
		msg := string(message)

		if strings.Contains(msg, "\"action\":\"connect\"") {
			if !sendMsg(c, "{action:connect,success:true}") {
				break
			} else {
				log.Println("{action:connect,success:true}", c)
			}
		}
		//log.Println("++++++++++++++++++++++++")
		if strings.Contains(msg, "\"action\":\"datasend\"") {
			//log.Println("-----------------------------")
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

				log.Println("incoming kotel data:", data)

				if !sendMsg(c, "{action:datasend,success:true}") {
					log.Println("Send to " + devId + ": failed")
					break
				} else {
					log.Println("Send to " + devId + ": success")
					err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}
					_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}

					//sendDataToUsers(msg);

				}

			}

			if strings.Contains(msg, "\"type\":\"roomdata\"") {
				var data ent.SensorsData

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

				data.DATE = time.Now();

				log.Println("incoming room data:", data)

				if !sendMsg(c, "{action:datasend,success:true}") {
					log.Println("Send to " + devId + ": failed")
					break
				} else {
					log.Println("Send to " + devId + ": success")
					success, err := db.AddRoomData(data)
					if err != nil || !success {
						log.Println("Error data writing in db: ", err)
					}
				}

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
		if strings.Contains(msg, "\"action\":\"getDestValues\"") {
			if strings.Contains(msg, "\"type\":\"koteldata\"") {
				kd, err := db.GetKotelData()
				log.Println("Get dest data from db: ", kd)
				if err != nil {
					log.Println("Error while read data from data base: ", err)
				}

				kData, err := json.Marshal(kd)
				if err != nil {
					log.Println("Error data marshaling: ", err)
				}

				sData := string(kData)
				msg := "{\"action\":\"setDestValues\"," + sData[66:len(sData)]
				//msg := strings.Replace(sData, "{", "{\"action\":\"setDestValues\", ", 1)
				//msg := "{\"action\":\"setDestValues\",\"device_id\":\"ESP_6822FC\",\"to\":65.35,\"tp\":50.34,\"kw\":11,\"pr\":2.43,\"destTo\":36.0,\"destTp\":29.0,\"destKw\":2,\"destPr\":1.8,\"destTc\":25.0,\"stage\":\"0_1\"}"
				log.Println("Dest Data is: ", msg)
				if !sendMsg(c, msg) {
					break
				}
			}

		}

	}
}



// writePump pumps messages from the hub to the websocket connection.
func (c *Conn) writePump(db db.DbService) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				//w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println("# Send ping to " + c.GetDeviceId() + " failed.  #")
				return
			} else {
				log.Println("# Send ping to " + c.GetDeviceId() + " success.  #")
			}
		}
	}
}

//func ping(appData string) error {
//	log.Println("[WS] Ping: "+appData)
//	return nil
//}
//
//func pong(appData string) error {
//	log.Println("[WS] Pong: "+appData)
//	return nil
//}

/*func wsProcessor(c *websocket.Conn, db db.DbService, dId string) {
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
			if !sendMsg(c, "{action:connect,success:true}") {
				break
			} else {
				log.Println("{action:connect,success:true}", c)
			}
		}
		//log.Println("++++++++++++++++++++++++")
		if strings.Contains(msg, "\"action\":\"datasend\"") {
			//log.Println("-----------------------------")
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

				log.Println("incoming kotel data:", data)

				if !sendMsg(c, "{action:datasend,success:true}") {
					log.Println("Send to " + devId + ": failed")
					break
				} else {
					log.Println("Send to " + devId + ": success")
					err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}
					_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
					if err != nil {
						log.Println("Error data writing in db: ", err)
					}

					sendDataToUsers(msg);

				}

			}

			if strings.Contains(msg, "\"type\":\"roomdata\"") {
				var data ent.SensorsData

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

				data.DATE = time.Now();

				log.Println("incoming room data:", data)

				if !sendMsg(c, "{action:datasend,success:true}") {
					log.Println("Send to " + devId + ": failed")
					break
				} else {
					log.Println("Send to " + devId + ": success")
					success, err := db.AddRoomData(data)
					if err != nil || !success {
						log.Println("Error data writing in db: ", err)
					}
				}

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
		if strings.Contains(msg, "\"action\":\"getDestValues\"") {
			if strings.Contains(msg, "\"type\":\"koteldata\"") {
				kd, err := db.GetKotelData()
				log.Println("Get dest data from db: ", kd)
				if err != nil {
					log.Println("Error while read data from data base: ", err)
				}

				kData, err := json.Marshal(kd)
				if err != nil {
					log.Println("Error data marshaling: ", err)
				}

				sData := string(kData)
				msg := "{\"action\":\"setDestValues\"," + sData[66:len(sData)]
				//msg := strings.Replace(sData, "{", "{\"action\":\"setDestValues\", ", 1)
				//msg := "{\"action\":\"setDestValues\",\"device_id\":\"ESP_6822FC\",\"to\":65.35,\"tp\":50.34,\"kw\":11,\"pr\":2.43,\"destTo\":36.0,\"destTp\":29.0,\"destKw\":2,\"destPr\":1.8,\"destTc\":25.0,\"stage\":\"0_1\"}"
				log.Println("Dest Data is: ", msg)
				if !sendMsg(c, msg) {
					break
				}
			}

		}
	}

}*/

func sendMsg(c *Conn, m string) bool {
	err := c.write(1, []byte(m))
	if err != nil {
		log.Println("[WS]:send:", err)
		return false
	}
	return true
}

//func sendMsg(c *websocket.Conn, m string) bool {
//	err := c.WriteMessage(1, []byte(m))
//	if err != nil {
//		log.Println("[WS]:send:", err)
//		return false
//	}
//	return true
//}

//func getDevIdByConn(c *websocket.Conn) string {
//	for k, v := range WsConnections {
//		if v == c {
//			return k
//		}
//	}
//
//	return ""
//}


//func pingActiveDevices() {
//	log.Println("########## Ping Devices ############")
//	for devId, c := range WsConnections {
//		if c != nil {
//			if !sendMsg(c, "ping") {
//				log.Println("# Send ping to " + devId + " failed.  #")
//				deleteWsConn(devId)
//			} else {
//				log.Println("# Send ping to " + devId + " success. #" )
//			}
//		} else {
//			log.Println("# Device " + devId + " is die.        #" )
//			deleteWsConn(devId)
//		}
//
//	}
//	log.Println("####################################")
//}

//func deleteWsConn(dId string) {
//	WsConnections[dId] = nil
//	delete(WsConnections, dId)
//}

//func sendDataToUsers(msg string) {
//	for devId, c := range WsConnections {
//		if strings.Contains(devId, "user_") {
//			if !sendMsg(c, msg) {
//				log.Println("# Send data to " + devId + " failed.  #")
//				deleteWsConn(devId)
//			} else {
//				log.Println("# Send data to " + devId + " success. #" )
//			}
//		}
//	}
//}