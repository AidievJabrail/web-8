package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

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

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") 
    ip := r.RemoteAddr
    name, _ := h.dbProvider.SelectUserSQL(ip)
    if name == "" {
        w.Write([]byte("Создайте пользователя через /api/user/post"))
        return
    }
    w.Write([]byte(fmt.Sprintf("Hello, %s!", name)))
}

func (h *Handlers) PostUser(w http.ResponseWriter, r *http.Request) {
    ip := r.RemoteAddr
	name := r.URL.Query().Get("name")
    err := h.dbProvider.AddUserSQL(name, ip)
    if err != nil{
        w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
        return
    }
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Запись добавлена!"))
}

func (dp *DatabaseProvider) SelectUserSQL(ip string) (string, error) {
	var resp string
	row := dp.db.QueryRow("SELECT name_user FROM users WHERE ip_address = $1", ip)
	err := row.Scan(&resp)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (dp *DatabaseProvider) FoundUserSQL(ip string) (bool) {
	var resp string
	row := dp.db.QueryRow("SELECT name_user FROM users WHERE ip_address = $1", ip)
	err := row.Scan(&resp)
	return err == nil
}

func (dp *DatabaseProvider) AddUserSQL(name, ip string) error {
    
    var err error
    if dp.FoundUserSQL(ip){
        _, err = dp.db.Exec("UPDATE users SET name_user = $1 WHERE ip_address = $2", name, ip)
    }else{
        _, err = dp.db.Exec("INSERT INTO users (name_user, ip_address) VALUES ($1, $2)", name, ip)
    }
	
	if err != nil {
		return err
	}

	return nil
}



func main() {

    address := flag.String("address", "127.0.0.1:8083", "адрес для запуска сервера")
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

    http.HandleFunc("/api/user", h.GetUser)
    http.HandleFunc("/api/user/post", h.PostUser)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
	
}