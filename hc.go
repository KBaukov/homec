package main

import (
	"github.com/KBaukov/homec/config"
	"github.com/KBaukov/homec/db"
	"github.com/KBaukov/homec/handle"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Listener chan net.Conn

var (
	//configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")
	Cfg = config.LoadConfig("config.json")
)

func init() { }

func main() {
	//flag.Parse()

	if Cfg.LoggerPath != "" {
		// Логер только добавляет данные
		logFile, err := os.OpenFile(Cfg.LoggerPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Ошибка открытия файла лога: %v", err)
		} else {
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}

	db, err := db.NewDB(Cfg.DbConnectionString)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных: %v", err)
		return;
	} else {
		db.Conn.SetMaxOpenConns(50);
		db.Conn.SetMaxIdleConns(5);
		stats := db.Conn.Stats().OpenConnections
		log.Printf("Open connections:", stats)
		//_, err = db.Conn.Exec("SET AUTOCOMMIT=1;")
		//if err != nil {
		//	log.Printf("Не удалось установить настройки базы данных: %v", err)
		//	return;
		//}
		defer db.Conn.Close();
	}

	//go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	frf := Cfg.FrontRoute.WebResFolder;

	http.HandleFunc("/logout", handle.ServeHome)
	http.HandleFunc("/login", handle.ServeLogin(db))
	http.HandleFunc("/home", handle.ServeHome)
	http.HandleFunc("/"+frf+"/", handle.ServeWebRes)
	http.HandleFunc("/css/", handle.ServeWebRes)
	http.HandleFunc("/js/", handle.ServeWebRes)
	http.HandleFunc("/img/", handle.ServeWebRes)
	http.HandleFunc("/fonts/", handle.ServeWebRes)
	http.HandleFunc("/api/", handle.ServeApi(db))
	http.HandleFunc("/ws", handle.ServeWs(db))

	//log.Fatal(http.ListenAndServe(*addr, nil))

	listenString := Cfg.Server.Address + ":" + Cfg.Server.Port
	log.Print("Сервер запущен: ", listenString)



	if Cfg.Server.TLS {
		err = http.ListenAndServeTLS(listenString, Cfg.Server.CertificatePath, Cfg.Server.KeyPath, nil)
	} else {
		err = http.ListenAndServe(listenString, nil)
	}
	if err != nil {
		log.Printf("Ошибка веб-сервера: %v", err)
	}

}

func redirect(w http.ResponseWriter, r *http.Request) {
	log.Print("Host: ", r.Host)
	listenString := "https://"+r.Host + ":" + Cfg.Server.Port+"/home"
	if len(r.URL.RawQuery) > 0 {
		listenString += "?" + r.URL.RawQuery
	}
	log.Printf("redirect to: %s", listenString)
	http.Redirect(w, r, listenString,
		// see @andreiavrammsd comment: often 307 > 301
		http.StatusTemporaryRedirect)
}
