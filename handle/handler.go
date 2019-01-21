package handle

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/KBaukov/homec/db"
	"github.com/KBaukov/homec/ent"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	sessStore = sessions.NewCookieStore([]byte("33446a9dcf9ea060a0a6532b166da32f304af0de"))
)

func init() {
	gob.Register(ent.User{})
	gob.Register(websocket.Conn{})

	sessStore.Options = &sessions.Options{
		Domain:   "*",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: false,
	}
}

type errData struct {
	Error_Code int
	Error_Message string
}


func ServeHome(w http.ResponseWriter, r *http.Request) {

	log.Println("Path: " + r.URL.Path)

	if r.URL.Path == "/logout" {
		session := getSession(w, r)
		//log.Println("++++Session: ", session)
		delete(session.Values, "user")
		session.Save(r, w)
		//log.Println("----Session: ", session)
		now := []byte(time.Now().String())
		sha := base64.URLEncoding.EncodeToString(now)
		http.Redirect(w, r, "/home?dc="+sha, 301)
		//http.ServeFile(w, r, "./webres/html/login.html")
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/home" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	} else {
		session := getSession(w, r)
		log.Println("Session: ", session)
		u:=session.Values["user"]
		if u == nil {
			log.Println("Redirect: /login")
			http.Redirect(w, r, "/login", 301)
		} else {
			user := u.(ent.User)
			t, err := template.ParseFiles("./webres/html/main.html")
			if err != nil {
				log.Println("Temlate parse error: ", err)
			}
			t.Execute(w, user)

		}
	}

}

func ServeLogin(db db.DbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" && r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Method == "POST" {

			login := strings.ToLower(r.PostFormValue("username"))
			pass := r.PostFormValue("password")

			pass, err := HashPass(pass)
			if err != nil {
				log.Println("Ошибка хеширования", err)
			}

			//log.Println("pass: ", pass)

			users, err := db.Auth(login, pass)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка авторизации (логин: %v): %v", login, err)
			}

			if len(users) != 1 {
				ed := errData{403, "Неправильные логин или пароль./nВ доступе отказано."}
				t, err := template.ParseFiles("./webres/html/login.html")
				if err != nil {
					log.Println("Temlate parse error: ", err)
				}
				t.Execute(w, ed)
				return
			}

			u := *users[0]
			log.Println("Login user: ", u)
			err = createSession(w, r, u, "user")
			if err != nil {

			}

			http.Redirect(w, r, "/home", 301)

		}
		if r.Method == "GET" {
			ed := errData{200, ""}
			t, err := template.ParseFiles("./webres/html/login.html")
			if err != nil {
				log.Println("Temlate parse error: ", err)
			}
			t.Execute(w, ed)
			//http.ServeFile(w, r, "./webres/html/login.html")
		}

	}

}


func ServeLogout(w http.ResponseWriter, r *http.Request) {
	session := getSession(w, r)
	log.Println("++++Session: ") //, session)
	session.Values["user"] = nil
	session.Save(r, w)
	log.Println("Session: ", session)

	http.Redirect(w, r, "/login", 301)

}

func ServeWebRes(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "webres") {
		http.ServeFile(w, r, "."+r.URL.Path)
	}
}

func ServeApi(db db.DbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("incoming request in: %v", r.URL.Path)

		session := getSession(w, r)
		u:=session.Values["user"]
		if u == nil {
			log.Println("Session die !!!")
			//apiDataResponse(w, []int{}, new error("hjhg"))
			http.Error(w, "Session die!!!", http.StatusUnauthorized);
			return;
		}

		//##############
		if r.URL.Path == "/api/devices" {
			devices, err := db.GetDevices()
			apiDataResponse(w, devices, err)
		}
		if r.URL.Path == "/api/device/edit" {
			id := r.PostFormValue("id")
			intId, err := strconv.Atoi(id)
			if err != nil {
				log.Println("err:", err.Error())
			}

			devType := r.PostFormValue("type")
			devName := r.PostFormValue("name")
			ip := r.PostFormValue("ip")
			actFlag := r.PostFormValue("active_flag")
			descr := r.PostFormValue("description")

			_, err = db.UpdDevice(intId, devType, devName, ip, actFlag, descr)
			apiDataResponse(w, []int{}, err)
		}
		if r.URL.Path == "/api/device/delete" {
			id := r.PostFormValue("id")
			intId, err := strconv.Atoi(id)
			if err != nil {
				log.Println("err:", err.Error())
			}

			_, err = db.DelDevice(intId)
			apiDataResponse(w, []int{}, err)
		}
		//###################
		if r.URL.Path == "/api/users" {
			users, err := db.GetUsers()
			apiDataResponse(w, users, err)
		}
		if r.URL.Path == "/api/user/edit" {
			id := r.PostFormValue("id")
			intId, err := strconv.Atoi(id)
			if err != nil {
				log.Println("err:", err.Error())
			}

			login := r.PostFormValue("login")
			pass := r.PostFormValue("pass")
			pass, err = HashPass(pass)
			if err != nil {
				log.Println("Ошибка хеширования", err)
			}
			usrType := r.PostFormValue("user_type")
			actFlag := r.PostFormValue("active_flag")
			lastVs := r.PostFormValue("last_visit")
			lastV, err := time.Parse("2006-01-02T00:00:00Z", lastVs)
			if err != nil {
				log.Println("date forma validation error:", err.Error())
			}

			_, err = db.UpdUser(intId, login, pass, usrType, actFlag, lastV)
			apiDataResponse(w, []int{}, err)
		}
		if r.URL.Path == "/api/user/delete" {
			id := r.PostFormValue("id")
			intId, err := strconv.Atoi(id)
			if err != nil {
				log.Println("err:", err.Error())
			}

			_, err = db.DelUser(intId)
			apiDataResponse(w, []int{}, err)
		}
		//##################
		if r.URL.Path == "/api/maps" {
			maps, err := db.GetMaps()
			apiDataResponse(w, maps, err)
		}
		if r.URL.Path == "/api/maps/edit" {
			ids := r.PostFormValue("id")
			id, err := strconv.Atoi(ids)
			if err != nil {
				log.Println("err:", err.Error())
			}
			title := r.PostFormValue("title")
			pict := r.PostFormValue("pict")

			ws := r.PostFormValue("w")
			wi, err := strconv.Atoi(ws)
			if err != nil {
				log.Println("err:", err.Error())
			}
			hs := r.PostFormValue("h")
			h, err := strconv.Atoi(hs)
			if err != nil {
				log.Println("err:", err.Error())
			}
			descr := r.PostFormValue("description")

			if err != nil {
				log.Println("date forma validation error:", err.Error())
			}

			_, err = db.UpdMap(id, title, pict, wi, h, descr)
			apiDataResponse(w, []int{}, err)
		}
		if r.URL.Path == "/api/maps/delete" {
			ids := r.PostFormValue("id")
			id, err := strconv.Atoi(ids)
			if err != nil {
				log.Println("err:", err.Error())
			}

			_, err = db.DelMap(id)
			apiDataResponse(w, []int{}, err)
		}
		//##############################################
		if r.URL.Path == "/api/sensors" {
			mapIds := r.PostFormValue("map_id")
			mapId, err := strconv.Atoi(mapIds)
			if err != nil {
				log.Println("err:", err.Error())
			} //////
			sens, err := db.GetMapSensors(mapId)
			apiDataResponse(w, sens, err)
		}
		if r.URL.Path == "/api/sensors/edit" {
			ids := r.PostFormValue("id")
			id, err := strconv.Atoi(ids)
			if err != nil {
				log.Println("err:", err.Error())
			} //////

			mapIds := r.PostFormValue("map_id")
			mapId, err := strconv.Atoi(mapIds)
			if err != nil {
				log.Println("err:", err.Error())
			} /////

			devIds := r.PostFormValue("device_id")
			devId, err := strconv.Atoi(devIds)
			if err != nil {
				log.Println("err:", err.Error())
			} /////

			xks := r.PostFormValue("xk")
			xk, err := strconv.ParseFloat(xks, 64)
			if err != nil {
				log.Println("err:", err.Error())
			} /////

			yks := r.PostFormValue("yk")
			yk, err := strconv.ParseFloat(yks, 64)
			if err != nil {
				log.Println("err:", err.Error())
			} /////

			sensorType := r.PostFormValue("type")
			pict := r.PostFormValue("pict")
			descr := r.PostFormValue("description")

			_, err = db.UpdMapSensor(id, mapId, devId, sensorType, xk, yk, pict, descr)
			apiDataResponse(w, []int{}, err)
		}
		if r.URL.Path == "/api/sensors/delete" {
			ids := r.PostFormValue("id")
			id, err := strconv.Atoi(ids)
			if err != nil {
				log.Println("err:", err.Error())
			}

			_, err = db.DelMapSensor(id)
			apiDataResponse(w, []int{}, err)
		}
		if r.URL.Path == "/api/sensors/lastid" {
			sens, err := db.GetLastId("map_sensors")
			apiDataResponse(w, sens, err)
		}
		if r.URL.Path == "/api/sensors/data" {
			devId := r.PostFormValue("device_id")
			data, err := db.GetRoomData(devId)
			apiDataResponse(w, data, err)
		}
		if r.URL.Path == "/api/sensors/stat" {
			devId := r.PostFormValue("device_id")
			from := r.PostFormValue("from")
			to := r.PostFormValue("to")
			count, _ := strconv.Atoi( r.PostFormValue("count") )
			data, err := db.GetRoomDataStat(devId, from, to, count)
			apiDataResponse(w, data, err)
		}

		//##############################################
		if r.URL.Path == "/api/kotel/getvalues" {
			data, err := db.GetKotelData()
			//log.Println("Values: ", data)
			apiDataResponse(w, data, err)
		}
		if r.URL.Path == "/api/kotel/setdest" {
			var (
				desttp float64 = 0.0;
				destto float64 = 0.0;
				desttc float64 = 0.0;
				destkw int = 0;
				destpr float64 = 0.0;
				stage string = ""
				err error
				msg       string
				kotelName string
			)

			p1 := r.PostFormValue("desttp");
			p2 := r.PostFormValue("destto");
			p3 := r.PostFormValue("destkw");
			p4 := r.PostFormValue("destpr");
			p5 := r.PostFormValue("desttc");
			p6 := r.PostFormValue("stage");

			if p1 != "" { desttp, _ = strconv.ParseFloat(p1, 64); }
			if p2 != "" { destto, _ = strconv.ParseFloat(p2, 64); }
			if p3 != "" { destkw, _ = strconv.Atoi(p3); }
			if p4 != "" { destpr, _ = strconv.ParseFloat(p4, 64); }
			if p5 != "" { desttc, _ = strconv.ParseFloat(p5, 64); }
			if p6 != "" { stage = p6; }

			kd, err := db.GetKotelData()

			if desttp == 0.0 {
				desttp = kd.DESTTP
			}
			if destto == 0.0 {
				destto = kd.DESTTO
			}
			if destkw == 0 {
				destkw = kd.DESTKW
			}
			if destpr == 0.0 {
				destpr = kd.DESTPR
			}
			if desttc == 0.0 {
				desttc = kd.DESTС
			}
			if stage == "" {
				stage = kd.STAGE
			}

			log.Println("Dest values: ", destto, desttp, destkw, destpr, desttc, stage)

			_, kotelName, err = db.GetKotelID()
			if err != nil || kotelName == "" {
				err = errors.New("Котел не найден")
			}
			ws := WsConnections[kotelName]
			if ws == nil {
				err = errors.New("Сессия не активна")
			} else {
				msg = "{\"action\":\"setDestValues\",\"destTo\":" + p2 + ",\"destTp\":" + p1 + ",\"destTc\":" +
					p5 + ",\"destPr\":" + p4 + ",\"destKw\":" + p3 + ",\"stage\":\"" + stage + "\"}";

				log.Printf("[WS]:send to %s: %s", kotelName, msg)
				err = ws.WriteMessage(1, []byte(msg))
				if err != nil {
					log.Println("Sending message error:", err)
				}
			}

			err = db.UpdKotelDestData(destto, desttp, destkw, destpr, desttc, stage)

			apiDataResponse(w, []int{}, err)

		}
		if r.URL.Path == "/api/kotel/getstage" {
			data, err := db.GetKotelData()
			apiDataResponse(w, data.STAGE, err)
		}
		if r.URL.Path == "/api/kotel/setstage" {
			stage := r.PostFormValue("stage");
			data, err := db.GetKotelData()
			data.STAGE = stage;

			apiDataResponse(w, data, err)
		}
		if r.URL.Path == "/api/kotel/pressbutt" {
			var (
				msg       string
				err       error
				kotelName string
			)

			_, kotelName, err = db.GetKotelID()
			if err != nil || kotelName == "" {
				err = errors.New("Котел не найден")
			}

			butt := r.PostFormValue("button")
			//stage := r.PostFormValue("stage")

			ws := WsConnections[kotelName]
			if ws == nil {
				err = errors.New("Сессия не активна")
			} else {
				msg = "{\"action\":\"pessButton\", \"butt\":\"" + butt +"\"}"; //"\", \"satge\": \""+stage+"\"}"
				log.Printf("Sending message to %s: %s", kotelName, msg)
				err = ws.WriteMessage(1, []byte(msg))
				if err != nil {
					log.Println("Sending message error:", err)
				} else {
					//for {
					//	if !WsPresButtFlag {
					//		time.Sleep(200 * time.Millisecond);
					//		break;
					//	}
					//}
				}
			}

			time.Sleep(300 * time.Millisecond);

			apiDataResponse(w, "", err)
		}
		if r.URL.Path == "/api/kotel/sessionstart" {
			var (
				msg       string
				err       error
				kotelName string
			)

			if IsControlSessionOpen {
				err =  errors.New("Котлом уже кто-то управляет")
			} else {

				_, kotelName, err = db.GetKotelID()
				if err != nil || kotelName == "" {
					err = errors.New("Котел не найден")
				}
				user := r.PostFormValue("user")
				ws := WsConnections[kotelName]
				if ws == nil {
					err = errors.New("Сессия не активна")
				} else {
					msg = "{\"action\":\"sessionStart\",\"user\":\"" + user + "\"}"
					log.Printf("Sending message to %s: %s", kotelName, msg)
					err = ws.WriteMessage(1, []byte(msg))
					if err != nil {
						log.Println("Sending message error:", err)
					}
				}
			}


			apiDataResponse(w, msg, err)
		}
		if r.URL.Path == "/api/kotel/sessionstop" {
			var (
				msg       string
				err       error
				kotelName string
			)
			_, kotelName, err = db.GetKotelID()
			if err != nil || kotelName == "" {
				err = errors.New("Котел не найден")
			}
			ws := WsConnections[kotelName]
			if ws == nil {
				err = errors.New("Сессия не активна")
			} else {
				msg = "{\"action\":\"sessionStop\"}"
				log.Printf("Sending message to %s: %s", kotelName, msg)
				err = ws.WriteMessage(1, []byte(msg))
				if err != nil {
					log.Println("Sending message error:", err)
				}
				IsControlSessionOpen = false;
			}

			apiDataResponse(w, msg, err)
		}
		if r.URL.Path == "/api/kotel/stat" {
			//devId := r.PostFormValue("device_id")
			from := r.PostFormValue("from")
			to := r.PostFormValue("to")
			count, _ := strconv.Atoi( r.PostFormValue("count") )
			data, err := db.GetKotelDataStat(from, to, count)
			apiDataResponse(w, data, err)
		}


		return
	}
}

func apiDataResponse(w http.ResponseWriter, data interface{}, err error) {
	errMsg := ""
	succes := true

	if err != nil {
		//http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка: %v", err)
		errMsg = err.Error()
		succes = false
	}

	dataResp := ent.ApiResp{SUCCESS: succes, DATA: data, MSG: errMsg}

	json, err := json.Marshal(dataResp)
	if err != nil {
		//http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		log.Printf("Ошибка маршалинга: %v", err)
		return
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	_, err = w.Write(json)
	if err != nil {
		log.Printf("Ошибка записи результата запроса: %v", err)
	}
}

//########################## helpers ############################

func HashPass(p string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(p))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func createSession(w http.ResponseWriter, r *http.Request, o interface{}, key string) error {

	session, err := sessStore.Get(r, "HomeControl")
	if err != nil {
		log.Printf("Error of session storage: %v", err)
		return err
	}

	session.Values[key] = o //User{"Pogi", "Points", ""}
	err = session.Save(r, w)
	if err!= nil {
		log.Printf("Error while create session: %v", err)
		return err
	}

	log.Println("Session created: succes")

	return err
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := sessStore.Get(r, "HomeControl")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		session, err = sessStore.New(r, "HomeControl")
	}
	return session
}

