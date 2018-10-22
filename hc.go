package main

import (
	"flag"
	"homec/config"
	"homec/db"
	"homec/handle"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Listener chan net.Conn

var (
	configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")
	cfg               = config.LoadConfig(*configurationPath)
)

func init() { }

func main() {
	flag.Parse()

	if cfg.LoggerPath != "" {
		// Логер только добавляет данные
		logFile, err := os.OpenFile(cfg.LoggerPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Ошибка открытия файла лога: %v", err)
		} else {
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}

	db, err := db.NewDB(cfg.DbConnectionString)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных: %v", err)
	} else {
		_, err = db.Conn.Exec("SET AUTOCOMMIT=1;")
		if err != nil {
			log.Printf("Не удалось установить настройки базы данных: %v", err)
		}
	}

	http.HandleFunc("/logout", handle.ServeHome)
	http.HandleFunc("/login", handle.ServeLogin(db))
	http.HandleFunc("/home", handle.ServeHome)
	http.HandleFunc("/webres/", handle.ServeWebRes)
	http.HandleFunc("/api/", handle.ServeApi(db))
	http.HandleFunc("/ws", handle.ServeWs(db))

	//log.Fatal(http.ListenAndServe(*addr, nil))

	listenString := cfg.Server.Address + ":" + cfg.Server.Port
	log.Print("Запуск сервера: ", listenString)

	if cfg.Server.TLS {
		err = http.ListenAndServeTLS(listenString, cfg.Server.CertificatePath, cfg.Server.KeyPath, nil)
	} else {
		err = http.ListenAndServe(listenString, nil)
	}
	if err != nil {
		log.Printf("Ошибка веб-сервера: %v", err)
	}

}
