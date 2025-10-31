package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleConnection)

}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		// add error handling
	}
	defer c.CloseNow()

	// Set the context as needed. Use of r.Context() is not recommended
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var v any
	err = wsjson.Read(ctx, c, &v)
	if err != nil {
		// add error handling
	}

	log.Printf("received: %v", v)

	c.Close(websocket.StatusNormalClosure, "")
}
