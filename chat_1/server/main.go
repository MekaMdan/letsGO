package main

import(
	"flag"
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)


type Message struct{

}

type hub struct {
	clients 
}