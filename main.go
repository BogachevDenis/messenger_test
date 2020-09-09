package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/messenger_test/database"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"
	"time"
)

type setting struct {
	ServerHost		string
	HTTPServerPort	string
	PgPort			string
	PgUser     		string
	PgPass     		string
	PgBase     		string
}


var num int
var c = make(chan database.Msg)
var cfg setting


func init()  {
	file, e := os.Open("setting.cfg")
	if e != nil{
		log.WithFields(log.Fields{
			"file" : "setting.cfg",
			"error" : e,
		}).Fatal("File can`t be opened")
	}
	log.WithFields(log.Fields{
		"file" : "setting.cfg",
	}).Info("Config file was opened")

	defer file.Close()
	stat, e := file.Stat()
	if e != nil{
		log.WithFields(log.Fields{
			"file" : "setting.cfg",
			"error" : e,
		}).Fatal("File can`t be stat")
	}

	readByte := make([]byte, stat.Size())
	_, error := file.Read(readByte)
	if error != nil{
		log.WithFields(log.Fields{
			"file" : "setting.cfg",
			"error" : error,
		}).Fatal("File can`t be read")
	}

	err := json.Unmarshal(readByte, &cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"file" : "cfg",
			"error" : err,
		}).Fatal("Config data can`t be unmarshal")
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/ws2", wsHandler2)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/client2", rootHandler2)
	http.HandleFunc("/client3", rootHandler3)
	log.WithFields(log.Fields{}).Info("starting server at :8080")
	panic(http.ListenAndServe(":"+cfg.HTTPServerPort, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	error_db := database.Connect(cfg.ServerHost,cfg.PgPort,cfg.PgUser, cfg.PgPass, cfg.PgBase)
	if error_db  != nil {
		log.WithFields(log.Fields{
			"database" : "msg",
			"error" : error_db,
		}).Fatal("Connection error")
	}
	log.WithFields(log.Fields{
		"database" : cfg.PgBase,
	}).Info("Database connection OK")
	tmpl := template.Must(template.ParseFiles("template/index.html"))
	tmpl.Execute(w, nil)

}
func rootHandler2(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/index2.html"))
	tmpl.Execute(w, nil)
}

func rootHandler3(w http.ResponseWriter, r *http.Request) {
	masMessage := database.SelectMessage()
	tmpl := template.Must(template.ParseFiles("template/index3.html"))
	tmpl.Execute(w, masMessage)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	go echo(conn, c)
}

func wsHandler2(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn2, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	go echo2(conn2, c)
}

func echo(conn *websocket.Conn,c chan database.Msg ) {
	for {
		msg := database.Msg{}
		t := time.Now()
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.WithFields(log.Fields{
				"error" : err,
			}).Fatal("Error reading json")
			break
		}
		time := t.Local()
		num++
		error := database.InsertMessage(num, msg.Message, time)
		if error != nil {
			log.WithFields(log.Fields{
				"InsertMessage" : "error",
				"error" : error,
			}).Fatal("Insert Database error")
		}
		msg.Num = num
		msg.Time = time.String()
		log.WithFields(log.Fields{
			"message" : msg,
		}).Info("Got message")

		c <- msg
		if err = conn.WriteJSON(msg); err != nil {
			log.Println(err)
		}
	}
}

func echo2(client *websocket.Conn, c chan database.Msg) {
	for {
		r := <-c
		w, err := client.NextWriter(websocket.TextMessage)
		if err != nil {
			break
		}
		data, _ := json.Marshal(r)
		msg := data
		w.Write(msg)
		log.WithFields(log.Fields{
			"message" : r,
		}).Info("Send message")
		w.Close()
	}
}

