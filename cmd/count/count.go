package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"encoding/json"
	"io"

	_ "github.com/lib/pq"
)
const (
	host     = "localhost"
	port     = 5432
	user     = "jabrail"
	password = "07012006"
	dbname   = "querydb"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

type myStruct struct {
	Count int `json:"count"`
}


func (h Handlers) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch r.Method {
	case http.MethodGet:
		count, err := h.dbProvider.GetCount()
		if err != nil{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
        	return
		}
		w.Write([]byte(strconv.Itoa(count)))
	case http.MethodPost:
		
		var tmp myStruct
		r.ParseForm()
		data, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(data, &tmp); err != nil {
			fmt.Println(err)
			return
		}
		err := h.dbProvider.IncrementCount(tmp.Count)
		if err != nil{
			w.Write([]byte(err.Error()))
        	return
		}
		w.Write([]byte("Успешно!"))
	}
}

func (h Handlers) SetCount(w http.ResponseWriter, r *http.Request){
	var tmp myStruct
	r.ParseForm()
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &tmp); err != nil {
		fmt.Println(err)
		return
	}
	err := h.dbProvider.SetCountSQL(tmp.Count)
	if err != nil{
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Успешно!"))
}

func (dp DatabaseProvider) IncrementCount(num int) error{
	_, err := dp.db.Exec("UPDATE countdb SET count = count + $1", num)
	if err != nil {
		return err
	}
	return nil
}

func (dp DatabaseProvider) SetCountSQL(num int) error{
	_, err := dp.db.Exec("UPDATE countdb SET count = $1", num)
	if err != nil {
		return err
	}
	return nil
}

func (dp DatabaseProvider) GetCount() (int, error){
	var resp int
	row := dp.db.QueryRow("SELECT count FROM countdb")
	err := row.Scan(&resp)
	if err != nil {
		return -100, err
	}
	return resp, nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}

	h := Handlers{dbProvider: dp}

    http.HandleFunc("/count", h.handler)
	http.HandleFunc("/count/set", h.SetCount)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}