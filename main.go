package main

/* incluir --novos-- pacotes no programa */
import(
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

//map where key is actually a pointer to WebSocket ->
//value is not needed but the map estruture is easier than a array
//to append and delete itens
var clients = make(map[*websocket.Conn]bool)

//Will act as a queue for messages sent by clients
var broadcast = make(chan Message)

// object with methods for taking a normal HTTP connection and upgrading to a WebSocket 
var upgrader = websocket.Upgrader{}

//defines message object
type Message struct {
	Email string 'json:"email"'
	Username string 'json:"username"'
	Message string 'json:"message"'
}

func main() {
	fs := http.FileServer(http.Dir("../public")) 
	http.Handle("/",fs)
}