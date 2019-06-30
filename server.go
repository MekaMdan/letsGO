package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"

	//"net"
	"net/http"
	"time"
)

//tipos
type Chat struct {
	clientes map[string]Cliente
	clientesLock sync.Mutex
	fila chan string
    id int
}

type Cliente struct {
	nome string
	conn *websocket.Conn
	sala *Chat
}


//chat
const NSALA = 5 //quantidade de salas

const BROADCASTDELAY = 100

var Sala [NSALA]Chat

func broadcastLoop(sala int) { //corotina do loop para dar broadcast nas mensagens
	for {
        Sala[sala].Broadcast()
		time.Sleep(BROADCASTDELAY * time.Millisecond)
	}
}

func (Sala *Chat) Initialize(sala int) { //inicializando a sala de chat
	Sala.clientes = make(map[string]Cliente)
	Sala.fila = make(chan string, 5)
    Sala.id = sala


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
	switch nome{
	case "schwarzenegger":
		Sala.InsereMsg("<font color='#CA0000'><I><B>" + nome + "</B> IS BACK</I></font>")
	case "cruzeiro":
		Sala.InsereMsg("<font color='#0058A2'><I><B>" + nome + "</B> campeão chegou</I></font>")
	case "ladeira":
		Sala.InsereMsg("<font color='#0058A2'><I><B>" + nome + "</B> chegou com a camisa do Cruzeiro </I></font>")
	default:
		Sala.InsereMsg("<B>" + nome + "</B> está entre nós.")
	}


    fmt.Printf("[SERVER] User %s - Sala %d (Conectado)\n", nome, Sala.id)

	return &cliente
}

//Logoff tira o usuario do chat
func (Sala *Chat) Logoff(nome string) { //faz logoff de um cliente do chat
    var message string
    Sala.clientesLock.Lock()
	delete(Sala.clientes, nome)
	Sala.clientesLock.Unlock()
	switch nome{
	case "schwarzenegger":
		message = "<font color='#CA0000'><I><B>" + nome + "</B> WILL BE BACK </I></font>"
	case "cruzeiro":
		message = "<font color='#0058A2'><I><B>" + nome + "</B> foi guardar os troféus </I><font>"
	case "ladeira":
		message = "<font color='#0058A2'><I><B>" + nome + "</B> foi assistir o jogo</I><font>"
	default:
		message = "<B>" + nome + "</B> não está mais entre nós."
	}

    Sala.InsereMsg(message)
    fmt.Printf("[SERVER] User %s - Sala %d (Desconectado)\n", nome, Sala.id)
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
    var message_form string

	switch User.nome {
	case "schwarzenegger":
		message_form = "<font color='#CA0000'><B><I>[" + t + "] " + User.nome + ":</B> " + msg + "</I></font>"
	case "cruzeiro":
		message_form = "<font color='#0058A2'><B><I>[" + t + "] " + User.nome + ":</B> " + msg + "</I></font>"
	case "ladeira":
		message_form = "<font color='#0058A2'><B><I>[" + t + "] " + User.nome + ":</B> " + msg + "</I></font>"
	default:
		message_form = "<B>[" + t + "] " + User.nome + ":</B> " + msg
	}

    User.sala.InsereMsg(message_form)
    fmt.Printf("[SALA %d] User %s - %s\n", User.sala.id, User.nome, msg)
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

func rotinaUser(conn *websocket.Conn) {
	_, msg, err := conn.ReadMessage()
    _, sala_temp, err2 := conn.ReadMessage()

    sala_string := string(sala_temp)
    sala, _ := strconv.Atoi(sala_string)

    cliente := Sala[sala].Login(string(msg), conn)
	if cliente == nil || err != nil  || err2 != nil {
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

func staticFiles(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./static"+r.URL.Path)
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

func webHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)

    if err != nil {
        println("Erro no websocket:", err)
        return
    }
    go rotinaUser(conn)
}

//Start  Printa a inicialização do client no terminal
func Start() {
    fmt.Printf("[SERVER] Servidor rodando...\n")
}

func main() {

	http.HandleFunc("/ws", webHandler)
	http.HandleFunc("/", staticFiles)
    for i:=0;i<NSALA;i++ {
        Sala[i].Initialize(i)
        go broadcastLoop(i)
    }
	Start()

	http.ListenAndServe(":8000", nil)
}
