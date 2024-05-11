package shared

import (
	"go-trivia/types"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	Lock     = &sync.RWMutex{}
	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	PlayerData      = make(map[string]types.Player)
	QuestionNumber  = 0
	BuzzChan        = make(chan bool)
	LeaderboardChan = make(chan bool)
	PlayersChan = make(chan bool)
)
