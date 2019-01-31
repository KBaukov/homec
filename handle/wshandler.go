package handle

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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
	maxMessageSize = 8192  				// Maximum message size allowed from peer.
	pingWait = 60 * time.Second 		// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second 		// Time allowed to read the next pong message from the peer.
	//pingPeriod = (pongWait * 9) / 10	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 1 * time.Second	// Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = 10 * time.Second	// Time to wait before force close on connection.
	brPref = "WBR_"
)

func init() {
	//log.Println("### Start Ping Scheduler ###")
	//gocron.Every(1).Second ().Do(pingActiveDevices)
	//gocron.Start()
}

//func internalError(ws *websocket.Conn, msg string, err error) {
//	log.Println(msg, err)
//	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
//}


var (
	Mu	sync.Mutex
	WsAsignConns	  = make(map[string]string)
	WsConnections	  = make(map[string] ent.WssConnect)
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
		var mu sync.Mutex

		deviceId := r.Header.Get("DeviceId")
		if deviceId =="" {
			suf,_ := HashPass(r.Header.Get("Sec-WebSocket-Key"));
			deviceId = brPref +strings.ToUpper(suf[12:18])
		}

		log.Println("incoming WS request from: ", deviceId)

		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		wsc := ent.WssConnect{mu, deviceId, ws,""}

		if(ws != nil) {
			ws.SetPingHandler(ping)
			ws.SetPongHandler(pong)
			//ws.SetReadDeadline(time.Now().Add(pongWait))
			ws.SetCloseHandler(wsClose)
			log.Println("Create new Ws Connection: succes, device: ", deviceId)
			Mu.Lock()
			WsConnections[deviceId] = wsc
			Mu.Unlock()
			go wsProcessor(&wsc, db)
		} else {
			log.Println("Ws Connectionfor device: ", deviceId, " not created.")
		}


	}
}

func ping(appData string) error {
	//log.Println("# Send ping to "+appData+" succes.     #")
	return nil
}

func pong(appData string) error {
	//log.Println("# Recive pong from "+appData+" succes. #")
	return nil
}

func wsClose(code int, reason string) error {
	log.Println("# Ws connection is closed: "+string(code)+" : "+ reason)
	return nil
}

func wsProcessor(wsc *ent.WssConnect, db db.DbService) {
	//var conn = wsc.Connection
	devId := wsc.DeviceId
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wsc.Connection.Close()
	}()

	for {

		select {
		case <-ticker.C:
			pingDevice(wsc, devId)
			log.Println("#########################################################################################################3")
		}

		msg, err := WsRead(wsc);
		if  err != nil {
			break
		}

		if strings.Contains(msg, "\"action\":\"connect\"") {
			if !sendMsg(wsc, "{\"action\":\"connect\",\"success\":true}") {
				break
			} else {
				log.Println("{action:connect,success:true}") //, conn)
			}
		}
		if strings.Contains(msg, "\"action\":\"assign\"") {
			assign := msg[29:len(msg)-2]
			if WsConnections[assign].Connection != nil {
				Mu.Lock()
				WsAsignConns[devId] = assign
				Mu.Unlock()
				log.Println("#### Assign", assign, " to ", devId)
				if !sendMsg(wsc, "{\"action\":\"assign\",\"success\":true}") {
					break
				}
			} else {
				if !sendMsg(wsc, "{\"action\":\"assign\",\"success\":false}") {
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

			rConn :=  getWsByDevId(rMsg.RECIPIENT)
			sender := devId //getDevIdByConn(c)

			message, _ := base64.URLEncoding.DecodeString(rMsg.MESSAGE)
			mm := strings.Replace(string(message),"\"sender\":\"\"", "\"sender\":\""+sender+"\"", 1)
			log.Println("[WS]:resend message: ",  mm, " to: ",  rMsg.RECIPIENT)
			sendMsg(&rConn, mm)
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

				if !sendMsg(wsc, "{action:datasend,success:true}") {
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

				if !sendMsg(wsc, "{action:datasend,success:true}") {
					break
				} else {
					success, err := db.AddRoomData(data)
					if err != nil || !success {
						log.Println("Error data writing in db: ", err)
					}
				}

				sendDataToWeb(msg, wsc.DeviceId);

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
				if !sendMsg(wsc, msg) {
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

func WsRead(wsc *ent.WssConnect) (string, error) {
	wsc.Mu.Lock()
	defer wsc.Mu.Unlock()

	if wsc.Connection != nil {
		mt, message, err := wsc.Connection.ReadMessage()
		if err != nil {
			log.Println("[WS]:error:  ", err)
			e := wsc.Connection.Close()
			if e!=nil {
				log.Println("[WS]:close connecton error:", err)
			} else {
				log.Println("[WS]:connection close...")
			}
			wsc.Connection =nil;
			deleteWsConn(wsc.DeviceId)
			return "", err
		}
		log.Printf("[WS]:recive: %s, type: %d", message, mt)
		return string(message), nil

	}  else {
		deleteWsConn(wsc.DeviceId)
		log.Println("[WS]:read from " + wsc.DeviceId + " failed")
		log.Println("[WS]:Connection lost: device - " + wsc.DeviceId)
		return "", errors.New("[WS]:read from " + wsc.DeviceId + " failed: connection lost")
	}


}


func sendMsg(wsc *ent.WssConnect, msg string) bool {
	wsc.Mu.Lock()
	defer wsc.Mu.Unlock();

	if wsc.Connection != nil {
		err := wsc.Connection.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("[WS]:send error:", err)
			e := wsc.Connection.Close()
			if e!=nil {
				log.Println("[WS]:close connecton error:", err)
			} else {
				log.Println("[WS]:connection close...")

			}
			wsc.Connection =nil;
			deleteWsConn(wsc.DeviceId)
			return false
		}
		log.Println("[WS]:send to " + wsc.DeviceId + " success: "+msg)
		return true
	} else {
		deleteWsConn(wsc.DeviceId)
		log.Println("[WS]:send to " + wsc.DeviceId + " failed: "+msg)
		log.Println("[WS]:Connection lost: device - " + wsc.DeviceId)
		return false
	}


}

func getDevIdByConn(c *websocket.Conn) string {
	for k, v := range WsConnections {
		if v.Connection == c {
			return k
		}
	}
	return ""
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

func pingActiveDevices() {
	log.Println("############ Ping Devices #############")
	for _, wsc := range WsConnections {
		devId := wsc.DeviceId
		if wsc.Connection != nil {
			if _, err := pingDevice(&wsc, devId); err != nil {
				log.Println("# Send ping to " + devId + " failed.     #")
				deleteWsConn(devId)
			} else {
				log.Println("# Send ping to " + devId + " success.    #" )
			}
		} else {
			log.Println("# Device " + devId + " is die.           #" )
			deleteWsConn(devId)
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

func deleteWsConn(dId string) {
	unassign(dId)
	//WsConnections[dId] = nil
	Mu.Lock()
	delete(WsConnections, dId)
	Mu.Unlock()
	log.Println("Delete connection #####################################################" )
}

func unassign(dId string) {
	for key, val := range WsAsignConns {
		if val == dId {
			wss := WsConnections[key]
			sendMsg(&wss, "{\"action\":\"unassign\",\"device\":\""+val+"\"}")
			Mu.Lock()
			delete(WsAsignConns, key)
			Mu.Unlock()
		}
	}
	Mu.Lock()
	delete(WsAsignConns, dId)
	Mu.Unlock()
	log.Println("Unassign connection #####################################################" )
}

func sendDataToWeb(msg string, sender string) {
	for devId, wsc := range WsConnections {
		if wsc.Connection==nil {
			deleteWsConn(devId);
			break;
		}
		assDev := wsc.DeviceId
		assRcp := WsAsignConns[assDev]
		if strings.Contains(devId, brPref) && sender==assRcp {
			if !sendMsg(&wsc, msg) {
				log.Println("# Send data to " + devId + " failed.  #")
				deleteWsConn(devId)
			} else {
				log.Println("# Send data to " + devId + " success. #" )
			}
		}
	}
}

