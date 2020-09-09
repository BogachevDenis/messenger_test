package database

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)
var db *sql.DB

type Msg struct {
	Num int
	Message string
	Time string
}

func Connect (pg_host, pg_port, pg_user, pg_pass, pg_base string ) error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",pg_host, pg_port, pg_user, pg_pass, pg_base))
	if err != nil {
		log.WithFields(log.Fields{
			"Connect" : "database",
			"error" : err,
		}).Fatal("Connection error")
	}
	return err
}

func InsertMessage(num int, msg string, time time.Time) error  {
	_, error := db.Exec(`INSERT INTO messages (num,message,time) VALUES($1,$2,$3)`,num,msg,time)
	if error != nil {
		log.WithFields(log.Fields{
			"Insert" : "messages",
			"error" : error,
		}).Fatal("Insert error")
	}
	return nil
}


func SelectMessage() []Msg {

	var arrm []Msg

	rows, err := db.Query(`SELECT "num","message","time" from "messages" WHERE time > now() - interval '1 hour'/3`)
	if err != nil {
		log.WithFields(log.Fields{
			"SelectMessage()" : "error",
			"error" : err,
		}).Fatal("Select error")
	}

	for rows.Next() {
		var m Msg
		e := rows.Scan(&m.Num,&m.Message,&m.Time)
		if e!=nil {
			log.WithFields(log.Fields{
				"Scan" : "rows",
				"error" : e,
			}).Fatal("Scan error")
		}
		arrm = append(arrm,m)

	}
	return arrm
}

