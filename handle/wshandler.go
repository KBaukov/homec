package handle

import (
	"encoding/base64"
	//"bytes"
	"encoding/json"
	"github.com/KBaukov/homec/db"
	"github.com/KBaukov/homec/ent"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 3 * time.Second 		// Time allowed to write a message to the peer.
	maxMessageSize = 1024  				// Maximum message size allowed from peer.
	pingWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second 		// Time allowed to read the next pong message from the peer.
	//pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 	1 * time.Second         // Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
	brPref = "WBR_"
)

func init() {
	//log.Println("### Start Ping Scheduler ###")
	//gocron.Every(1).Minute().Do(pingActiveDevices)
	//gocron.Start()
}

//func internalError(ws *websocket.Conn, msg string, err error) {
//	log.Println(msg, err)
//	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
//}


var (
	Mu					sync.Mutex
	WsAsignConns	  = make(map[string]string)
	WsConnections	  = make(map[string] ent.WssConnect)
	WsAllowedOrigin   = "HomeControlApp" //homec.Cfg.WsAllowedOrigin
	IsControlSessionOpen = false;
	WsPresButtFlag = false;
	WsAsignConns	  = make(map[string]string)
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
		var mu sync.Mutex

		if devId =="" {
			suf,_ := HashPass(r.Header.Get("Sec-WebSocket-Key"));
			suf,_ := HashPass(r.Header.Get("Sec-WebSocket-Key"));
			deviceId = brPref +strings.ToUpper(suf[12:18])
		}

		log.Println("incoming WS request from: ", devId)

		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		wsc := ent.WssConnect{mu, deviceId, ws,""}
		Mu.Lock()
		WsConnections[deviceId] = wsc
		Mu.Unlock()

		//WsConnections[deviceId] = ws
		//if(ws != nil) {
		//	//ws.SetPingHandler(ping)
			//ws.SetReadDeadline(time.Now().Add(pongWait))
			ws.SetCloseHandler(wsClose)
		//	go wsProcessor(ws, db, deviceId)
			go wsProcessor(&wsc, db)
		//	log.Println("Ws Connectionfor device: ", deviceId, " not created.")
		//}


	}
}

/*func ping(appData string) error {
	//log.Println("# Send ping to "+appData+" succes.     #")
	defer func() {
		hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	devId := c.deviceId
	//log.Println("# Recive pong from "+appData+" succes. #")
	return nil
}*/

func wsClose(code int, reason string) error {
	log.Println("# Ws connection is closed: "+string(code)+" : "+ reason)

func wsProcessor(wsc *ent.WssConnect, db db.DbService) {
	//var conn = wsc.Connection
	var devId = wsc.DeviceId
	defer wsc.Connection.Close()

	for {
		Mu.Lock()
		msg, err := WsRead(wsc);
		Mu.Unlock()
		if  err != nil {
				}
				break
			}


			if strings.Contains(msg, "\"action\":\"connect\"") {
			Mu.Lock()
			if !sendMsg(wsc, "{\"action\":\"connect\",\"success\":true}") {
				Mu.Unlock()
					break
				} else {
				log.Println("{action:connect,success:true}") //, conn)
				}
			Mu.Unlock()
			}
		if strings.Contains(msg, "\"action\":\"assign\"") {
			assign := msg[29:len(msg)-2]
			Mu.Lock()
			if WsConnections[assign].Connection != nil {
				WsAsignConns[devId] = assign
				log.Println("#### Assign", assign, " to ", devId)
				if !sendMsg(wsc, "{\"action\":\"assign\",\"success\":true}") {
					Mu.Unlock()
					break
				}

			} else {
				if !sendMsg(wsc, "{\"action\":\"assign\",\"success\":false}") {
					Mu.Unlock()
					break
				}
			}
			Mu.Unlock()

		}

		if strings.Contains(msg, "\"action\":\"resend\"") {
			var rMsg ent.ResendMessage

			err = json.Unmarshal([]byte(msg), &rMsg)
			if err != nil {
				log.Println("Error data unmarshaling: ", err)
			}

			rConn :=  getWsByDevId(rMsg.RECIPIENT)
			sender := devId //getDevIdByConn(c)

			message, _ := base64.URLEncoding.DecodeString(rMsg.MESSAGE)
			mm := strings.Replace(string(message),"\"sender\":\"\"", "\"sender\":\""+sender+"\"", 1)
			log.Println("[WS]:resend message: ",  mm, " to: ",  rMsg.RECIPIENT)
			Mu.Lock()
			sendMsg(&rConn, mm)
			Mu.Unlock()
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
				Mu.Lock()
				if !sendMsg(wsc, "{action:datasend,success:true}") {
					Mu.Unlock()
						break
					} else {

					err = db.UpdKotelMeshData(data.TO, data.TP, data.KW, data.PR, data.STAGE)
						if err != nil {
							log.Println("Error data writing in db: ", err)
						}
						_, err = db.AddKotelStatData(data.TO, data.TP, data.KW, data.PR)
						if err != nil {
							log.Println("Error data writing in db: ", err)
						}

					sendDataToWeb(msg, devId);

					}
				Mu.Unlock()
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
				Mu.Lock()
				if !sendMsg(wsc, "{action:datasend,success:true}") {
					Mu.Unlock()
						break
					} else {
						success, err := db.AddRoomData(data)
						if err != nil || !success {
							log.Println("Error data writing in db: ", err)
						}
					}

				sendDataToWeb(msg, wsc.DeviceId);
				Mu.Unlock()
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
				Mu.Lock()
				if !sendMsg(wsc, msg) {
					Mu.Unlock()
						break
					}
				Mu.Unlock()
				}
			if strings.Contains(msg, "\"type\":\"roomdata\"") {

			}

		} else {
			log.Println("# Connection lost    #")
			hub.unregister <- c
			break
		}
		if strings.Contains(msg, "\"action\":\"getMeshValues\"") {
			if strings.Contains(msg, "\"type\":\"koteldata\"") {

			}
		}
	}

}

func WsRead(wsc *ent.WssConnect) (string, error) {
	wsc.Mu.Lock()
	defer wsc.Mu.Unlock()

	mt, message, err := wsc.Connection.ReadMessage()
	if err != nil {
		log.Println("[WS]:error:  ", err)
		return "", err
	}
	log.Printf("[WS]:recive: %s, type: %d", message, mt)

	return string(message), nil
}


func sendMsg(wsc *ent.WssConnect, msg string) bool {
	wsc.Mu.Lock()
	defer wsc.Mu.Unlock();

	err := wsc.Connection.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					return
				}
	//devId := getDevIdByConn(c)
	log.Println("[WS]:send to " + wsc.DeviceId + " success: "+msg)
				n := len(c.send)
				for i := 0; i < n; i++ {
					//w.Write(newline)
					w.Write(<-c.send)
				}

				if err := w.Close(); err != nil {
					return
		if v.Connection == c {
				}
			case <-ticker.C:
				//log.Println("############################ Ping ##############################")
					if err := c.write(websocket.PingMessage, []byte{}); err != nil {
						log.Println("# Send ping to " + c.GetDeviceId() + " failed.  #")
						return
					} else {
						log.Println("# Send ping to " + c.GetDeviceId() + " success.  #")
					}

func getWsByConn(c *websocket.Conn) ent.WssConnect {
	for _, v := range WsConnections {
		if v.Connection == c {
			return v;
		}
	}
	return ent.WssConnect{};
}

func getWsByDevId(devId string) ent.WssConnect {
	for _, v := range WsConnections {
		if v.DeviceId == devId {
			return v;
		}
	}
	return ent.WssConnect{};
}

	log.Println("############ Ping Devices #############")
	for _, wsc := range WsConnections {
		devId := wsc.DeviceId
		if wsc.Connection != nil {
			if _, err := pingDevice(&wsc, devId); err != nil {
				log.Println("# Send ping to " + devId + " failed.     #")
				log.Println("# Send ping to " + devId + " success.    #" )
			}
		} else {
			log.Println("# Device " + devId + " is die.           #" )
			log.Println("# Connection lost    #")
		}
	}
	//log.Println("####################################")
}
func pingDevice(wsc *ent.WssConnect, devId string) (bool,error) {
	wsc.Mu.Lock();
	defer wsc.Mu.Unlock()
	err := wsc.Connection.WriteMessage(websocket.PingMessage, []byte(devId));
	if err != nil {
		return false, err
	}
	return true, err;
}

func sendMsg(c *Conn, m string) bool {
	unassign(dId)
	//WsConnections[dId] = nil
		log.Println("[WS]:send:", m, " succes")
		return true
	} else {
		log.Println("[WS]:send:", m, " failed")
		return false
	}

func unassign(dId string) {
	for key, val := range WsAsignConns {
		if val == dId {
			wss := WsConnections[key]
			sendMsg(&wss, "{\"action\":\"unassign\",\"device\":\""+val+"\"}")
			delete(WsAsignConns, key)
		}
	}
	delete(WsAsignConns, dId)
}

func sendDataToWeb(msg string, sender string) {
	for devId, wsc := range WsConnections {
		if wsc.Connection==nil {
			deleteWsConn(devId);
			break;
		}
		assDev := wsc.DeviceId//getDevIdByConn(c.Connection)
		assRcp := WsAsignConns[assDev]
		if strings.Contains(devId, brPref) && sender==assRcp {
			if !sendMsg(&wsc, msg) {
				//deleteWsConn(devId)
			}
			delete(WsAsignConns, key)
		}
	}
}

