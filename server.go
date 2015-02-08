package main

import (
	"github.com/beatrichartz/martini-sockets"
	"github.com/go-martini/martini"
	"log"
	"net/http"
)

func main() {
	app := martini.Classic()

	static := martini.Static("assets")
	app.NotFound(static, http.NotFound)

	gamesNeedingPlayer := []*Game{}
	activeGames := make(map[int]*Game)

	gameCount := 0

	app.Get("/sockets", sockets.JSON(Message{}), func(params martini.Params,
		receiver <-chan *Message,
		sender chan<- *Message,
		done <-chan bool,
		disconnect chan<- int,
		err <-chan error) (int, string) {

		player := &Player{0, false, receiver, sender, done, err, disconnect}

		if len(gamesNeedingPlayer) > 0 {
			game := gamesNeedingPlayer[0]
			gamesNeedingPlayer = gamesNeedingPlayer[1:]
			log.Printf("SERVER: player joining existing game %d.\n", game.Id)
			activeGames[game.Id] = game
			player.GameId = game.Id
			game.Join(player)
			log.Printf("SERVER: %d games needing players left.\n", len(gamesNeedingPlayer))
			log.Printf("SERVER: %d active games.\n", len(activeGames))
		} else {
			gameCount += 1
			player.GameId = gameCount
			gamesNeedingPlayer = append(gamesNeedingPlayer, &Game{gameCount, false, player, nil, [3][3]int{}})
			log.Printf("SERVER: player joining new game  %d.\n", player.GameId)
		}

		for {
			select {
			case msg := <-receiver:
				log.Println(msg)
				if msg.Type == -2 {
					game := activeGames[player.GameId]
					log.Printf("SERVER: game %d executing move\n", game.Id)
					game.DoMove(msg.X, msg.Y)
				}
			case <-done:
				game, prs := activeGames[player.GameId]
				if prs {
					log.Printf("SERVER: game %d ending\n", game.Id)
					game.End()
					delete(activeGames, player.GameId)
					log.Printf("SERVER: %s games left\n", len(activeGames))
				}
				return 200, "OK"
			}
		}
	})

	app.Run()
}
