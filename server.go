package main

import (
    "fmt"
    "./websocket"
    "sync"
    "net"
    "net/http"
)


//chat
type Chat struct{
    clientes map[string]Cliente
    clientesLock  sync.Mutex
}

type Cliente struct{
    nome string
    conn *websocket.Conn
    sala *Chat
}

var Sala Chat

func broadcastLoop (seconds int){   //corotina do loop para dar broadcast nas mensagens
    for {
        Sala.Broadcast()
        time.Sleep(seconds * time.Millisecond)
    }
}

func (Sala *Chat) Initialize(){ //inicializando a sala de chat
    Sala.clientes = make(map[string]Cliente)

    go broadcastLoop(100)

}


func (Sala *Chat) Login(nome string, conn *websocket.Conn) *Cliente{    //faz login de um cliente do chat
    defer Sala.clientesLock.Unlock();

    Sala.clientesLock.Lock();

    if _, exists := cr.clients[name]; exists {
		return nil
	}

    cliente := Cliente{
        nome: nome,
        conn: conn,
        sala: Sala,
    }

    Sala.clientes[nome] = cliente;
    if(nome = "schwarzenegger"){
        Sala.InsereMsg("<B>" + nome + "</B> IS BACK!")
    }
    else{
        Sala.InsereMsg("<B>" + nome + "</B> esta entre nos.")
    }
}


func (Sala *Chat) Logoff(nome string){  //faz logoff de um cliente do chat
    Sala.clientesLock.Lock();
    delete(Sala.clientes, nome)
}

func main(){

}
