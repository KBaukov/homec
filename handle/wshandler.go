package handle

import (
	"encoding/base64"
	"github.com/KBaukov/homec/config"

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
	//writeWait = 10 * time.Second 		// Time allowed to write a message to the peer.
//	//maxMessageSize = 1024  				// Maximum message size allowed from peer.
//	//pingWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
//	//pongWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
//	////pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
//	//pingPeriod = 	1 * time.Second         // Send pings to peer with this period. Must be less than pongWait.
//	//closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
//	////brPref = "WBR_"
)

func init() {
	cfg := config.LoadConfig("config.json")
	WsConfig := cfg.WsConfig

	writeWait = time.Duration(WsConfig.WriteWait) * time.Second                 // Time allowed to write a message to the peer.
	maxMessageSize = WsConfig.MaxMessageSize                                    // Maximum message size allowed from peer.
	pingWait = time.Duration(WsConfig.PingWait) * time.Second                   // Time allowed to read the next pong message from the peer.
	pongWait = time.Duration(WsConfig.PongWait) * time.Second                   // Time allowed to read the next pong message from the peer.
	pingPeriod = time.Duration(WsConfig.PingPeriod) * time.Second               // Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = time.Duration(WsConfig.CloseGracePeriod) * time.Second	// Time to wait before force close on connection.
	brPref = WsConfig.BrPref
	wsAllowedOrigin = WsConfig.WsAllowedOrigin

	log.Println("Config values: ", cfg)

	go hub.run()
}

var (

	writeWait time.Duration
	maxMessageSize int64
	pingWait time.Duration
	pongWait time.Duration
	pingPeriod time.Duration
	closeGracePeriod time.Duration
	brPref string
	wsAllowedOrigin   string
	isControlSessionOpen = false;
	WsAsignConns	  = make(map[string]string)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		var user = ent.User{}
		session := getSession(nil, r)
		u:=session.Values["user"]
		if u != nil {
			user = u.(ent.User)
		}
		deviceOrigin := r.Header.Get("Origin")
		token := r.Header.Get("Sec-WebSocket-Protocol")
		if wsAllowedOrigin == deviceOrigin || token == user.PASS[10:16]{
			return true
		} else {
			log.Println("Origin not allowed:", deviceOrigin)
			return false
		}

	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Println("{WS}:Error:", reason.Error())
	},
}

func ServeWs(db db.DbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var ws *websocket.Conn
		var err error

		devId := r.Header.Get("DeviceId")
		if devId =="" {
			suf,_ := HashPass(r.Header.Get("Sec-WebSocket-Key"));
			devId = brPref +strings.ToUpper(suf[12:18])
		}

		log.Println("incoming WS request from: ", devId)

		swp := r.Header.Get("Sec-WebSocket-Protocol")
		if swp!="" {
			headers := http.Header{"Sec-Websocket-Protocol": {swp}}
			ws, err = upgrader.Upgrade(w, r, headers)
		} else {
			ws, err = upgrader.Upgrade(w, r, nil)
		}

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
		if c != nil {
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
					log.Println("{action:connect,success:true}", c.deviceId)
				}
			}

			if strings.Contains(msg, "\"action\":\"assign\"") {
				assign := msg[29:len(msg)-2]
				cc := hub.getConnByDevId(assign)
				if cc != nil {
					WsAsignConns[devId] = assign
					log.Println("#### Assign", assign, " to ", devId)
					if !sendMsg(c, "{\"action\":\"assign\",\"success\":true}") {
						break
					}
				} else {
					if !sendMsg(c, "{\"action\":\"assign\",\"success\":false}") {
						break
					}
				}

			}

			if strings.Contains(msg, "\"action\":\"resend\"") {
				var rMsg ent.ResendMessage

				err = json.Unmarshal([]byte(msg), &rMsg)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				rConn :=  hub.getConnByDevId(rMsg.RECIPIENT)
				sender := devId

				message, _ := base64.URLEncoding.DecodeString(rMsg.MESSAGE)
				mm := strings.Replace(string(message),"\"sender\":\"\"", "\"sender\":\""+sender+"\"", 1)
				log.Println("[WS]:resend message: ",  mm, " to: ",  rMsg.RECIPIENT)
				sendMsg(rConn, mm)
			}

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

					//log.Println("incoming kotel data:", data)

					if !sendMsg(c, "{action:datasend,success:true}") {
						log.Println("Send to " + devId + ": failed")
						break
					} else {
						//log.Println("Send to " + devId + ": success")
						err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR)
						if err != nil {
							log.Println("Error data writing in db: ", err)
						}
						_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
						if err != nil {
							log.Println("Error data writing in db: ", err)
						}

						hub.sendDataToWeb(msg, devId);

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

					//log.Println("incoming room data:", data)

					if !sendMsg(c, "{action:datasend,success:true}") {
						log.Println("Send to " + devId + ": failed")
						break
					} else {
						//log.Println("Send to " + devId + ": success")
						success, err := db.AddRoomData(data)
						if err != nil || !success {
							log.Println("Error data writing in db: ", err)
						}
					}

					hub.sendDataToWeb(msg, devId);

				}

				if strings.Contains(msg, "\"type\":\"floordata\"") {
					var data ent.FloorData

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

					log.Println("incoming floor data:", data)

					if !sendMsg(c, "{action:datasend,success:true}") {
						log.Println("Send to " + devId + ": failed")
						break
					} else {
						//log.Println("Send to " + devId + ": success")
						err = db.UpdFloorMeshData(data)
						if err != nil {
							log.Println("Error data writing in db: ", err)
						}
						//_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
						//if err != nil {
						//	log.Println("Error data writing in db: ", err)
						//}

						hub.sendDataToWeb(msg, devId);

					}
				}
			}
			if strings.Contains(msg, "\"action\":\"pessButton\"") {
				if strings.Contains(msg, "true") {
					//WsPresButtFlag = false;
				}

			}
			if strings.Contains(msg, "\"action\":\"sessionStart\"") {
				if strings.Contains(msg, "true") {
					isControlSessionOpen = true;
				}

			}
			if strings.Contains(msg, "\"action\":\"sessionStop\"") {
				if strings.Contains(msg, "true") {
					isControlSessionOpen = false;
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
				if strings.Contains(msg, "\"type\":\"floordata\"") {

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
				if strings.Contains(msg, "\"type\":\"floordata\"") {

				}

			}

		} else {
			log.Println("# Connection lost    #")
			hub.unregister <- c
			break
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
		if c != nil {
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
				//log.Println("############################ Ping ##############################")

					if err := c.write(websocket.PingMessage, []byte{}); err != nil {
						log.Println("# Send ping to " + c.GetDeviceId() + " failed.  #")
						return
					} else {
						log.Println("# Send ping to " + c.GetDeviceId() + " success.  #")
					}


			}
		} else {
			hub.unregister <- c
			log.Println("# Connection lost    #")
		}
	}
}

func sendMsg(c *Conn, m string) bool {
	if c != nil {
		c.send <- []byte(m)
		log.Println("[WS]:send:", m, " succes")
		return true
	} else {
		log.Println("[WS]:send:", m, " failed")
		return false
	}


}

func unAssign(devId string) {
	for key, val := range WsAsignConns {
		if val == devId {
			conn := hub.getConnByDevId(key);
			if conn != nil {
				hub.getConnByDevId(key).send <- []byte("{\"action\":\"unassign\",\"device\":\""+val+"\"}")
			}
			delete(WsAsignConns, key)
		}
	}
}
