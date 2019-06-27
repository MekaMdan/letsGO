package main

import (
    "fmt"
    "github.com/gorilla/websocket"
	"sync"
	"strconv"
    //"net"
    "net/http"
    "time"
)

	//"net"
	"net/http"
	"time"
)

//tipos
type Chat struct {
	clientes     map[string]Cliente
	clientesLock sync.Mutex
	fila         chan string
}

type Cliente struct {
	nome string
	conn *websocket.Conn
	sala *Chat
}

//chat
var Sala Chat

func broadcastLoop() { //corotina do loop para dar broadcast nas mensagens
	for {
		Sala.Broadcast()
		time.Sleep(100 * time.Millisecond)
	}
}

func (Sala *Chat) Initialize() { //inicializando a sala de chat
	Sala.clientes = make(map[string]Cliente)
	Sala.fila = make(chan string, 5)

	go broadcastLoop()

}

//Login realiza o login de um cliente
func (Sala *Chat) Login(nome string, conn *websocket.Conn) *Cliente { //faz login de um cliente do chat
	//Conn é a conexão do websocket

	defer Sala.clientesLock.Unlock() //Só é executado depois q a função finaliza

	Sala.clientesLock.Lock()

	if _, exists := Sala.clientes[nome]; exists {
		return nil
	}

	cliente := Cliente{
		nome: nome,
		conn: conn,
		sala: Sala,
	}

	Sala.clientes[nome] = cliente
	if nome == "schwarzenegger" {
		Sala.InsereMsg("<I><B>" + nome + "</B> IS BACK</I>")
	} else {
		Sala.InsereMsg("<B>" + nome + "</B> está entre nós.")
	}
	return &cliente
}

//Logoff tira o usuario do chat
func (Sala *Chat) Logoff(nome string) { //faz logoff de um cliente do chat
	Sala.clientesLock.Lock()
	delete(Sala.clientes, nome)
	Sala.clientesLock.Unlock()
	if nome == "schwarzenegger" {
		Sala.InsereMsg("<I><B>" + nome + "</B> WILL BE BACK </I>")
	} else {
		Sala.InsereMsg("<B>" + nome + "</B> não está mais entre nós.")
	}
}

//InsereMsg Insere mensagem na fila do broadcast
func (Sala *Chat) InsereMsg(msg string) { //insere mensagem, na fila para broadcast
	Sala.fila <- msg
}

//Broadcast fica fazendo o loop de receber a mensagem e imprimir
func (Sala *Chat) Broadcast() {
	bloco := ""
loop:
	for {
		select {
		case temp := <-Sala.fila:
			bloco += temp + "<BR>"
		default:
			break loop
		}
	}
	if len(bloco) > 0 {
		for _, cliente := range Sala.clientes {
			cliente.Enviar(bloco)
		}
	}
}

func getHoras() string {
	hour := time.Now().Hour()
	min := time.Now().Minute()
	thour := strconv.Itoa(hour)
	tmin := strconv.Itoa(min)
	if hour < 10 {
		thour = "0" + thour
	}
	if min < 10 {
		tmin = "0" + tmin
	}
	return thour + ":" + tmin
}

//NovaMsg quando o cliente quer enviar uma mensagem
func (User *Cliente) NovaMsg(msg string) { //quer mandar uma mensagem
	t := getHoras()
	if User.nome == "schwarzenegger" {
		User.sala.InsereMsg("<B><I>[" + t + "] " + User.nome + ":</B> " + msg + "</I>")
	} else {
		User.sala.InsereMsg("<B>[" + t + "] " + User.nome + ":</B> " + msg)
	}
}

//Sair realiza a saida do cliente
func (User *Cliente) Sair() { //cliente quer sair
	User.sala.Logoff(User.nome)
}

//Enviar Envia a mensagem do usuário
func (User *Cliente) Enviar(msgs string) {
	User.conn.WriteMessage(websocket.TextMessage, []byte(msgs))
}

//PARTE DE HTML - PAGINA DA WEB DO LADO DO CLIENTE
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static"+r.URL.Path)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func rotinaUser(conn *websocket.Conn) {
	_, msg, err := conn.ReadMessage()
	cliente := Sala.Login(string(msg), conn)
	if cliente == nil || err != nil {
		conn.Close() //em caso de erro na leitura da mensagem ou no cliente, fecha a conexao e retorna a corotina
		return
	}

	for { //cliente espera por mensagens
		_, msg, err := conn.ReadMessage()
		if err != nil {
			cliente.Sair()
			return
		}
		cliente.NovaMsg(string(msg))
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Erro no websocket:", err)
		return
	}
	go rotinaUser(conn)
}

//Start  Printa a inicialização do client no terminal
func Start() {
	println("Servidor rodando...")
}

func main() {
	http.HandleFunc("/ws", webHandler)
	http.HandleFunc("/", staticFiles)
	Sala.Initialize()
	Start()
	http.ListenAndServe(":8000", nil)
}
