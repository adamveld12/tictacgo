package main

import (
	"fmt"
	"log"
)

const (
	Negotiating = iota
	Player1Turn
	Player2turn
	Over
)

type Game struct {
	Id      int
	Over    bool
	Player1 *Player
	Player2 *Player
	// -1 is 0
	// 1 is X
	Board [3][3]int
}

func (g *Game) Join(player *Player) {
	g.Player2 = player
	// now we start the game
	g.Player2.Negotiating()
	g.Player1.Negotiating()

	g.Player1.StartTurn()
	g.Player2.EndTurn()
	g.Over = false
}

func (g *Game) DoMove(x, y int) {
	if g.Over {
		return
	}

	// 1 == x
	move := 1
	player := g.Player1
	other := g.Player2
	if g.Player2.CurrentTurn {
		player = g.Player2
		other = g.Player1
		// -1 == o
		move = -1
	}
	log.Println("SERVER: Doing move for player")

	if !g.ValidMove(x, y) {
		log.Println("SERVER: Invalid move")
		player.UpdateStatus("That is not a valid move, try again.")
	} else {
		// send the move to both players
		g.Board[x][y] = move
		other.SendMove(x, y, move, "Opponent", move == 1)
		player.SendMove(x, y, move, "You", move == 1)
		log.Printf("SERVER: game %d player played (%d,%d)\n", g.Id, x, y)

		// err is a tie
		if win, p := g.CheckWin(); win {
			g.Over = true
			if p > 0 {
				log.Println("Player 1 (x) won!")
				g.Player1.EndGame(true)
				g.Player2.EndGame(false)
			} else if p < 0 {
				log.Println("Player 2 (o) won!")
				g.Player1.EndGame(false)
				g.Player2.EndGame(true)
			} else {
				log.Println("Tie")
				g.Player1.EndGame(false)
				g.Player2.EndGame(false)
			}
		} else {
			player.EndTurn()
			other.StartTurn()
		}
	}
}

func (g *Game) ValidMove(x, y int) bool {
	return x < 3 && x >= 0 && y < 3 && y >= 0 && g.Board[x][y] == 0
}

// True if board is in end state.
// Second return value is who won with 1 being x, 0 being tie and -1 being 0
func (g *Game) CheckWin() (bool, int) {
	b := g.Board
	xWins, oWins, movesLeft := false, false, true

	// checks 3 consecutive in row and column
	for i := 0; i < 3; i++ {
		xWins = (b[i][0] > 0 && b[i][1] > 0 && b[i][2] > 0) || (b[0][i] > 0 && b[1][i] > 0 && b[2][i] > 0) || xWins
		oWins = (b[i][0] < 0 && b[i][1] < 0 && b[i][2] < 0) || (b[0][i] < 0 && b[1][i] < 0 && b[2][i] < 0) || oWins
		if xWins || oWins {
			break
		}
	}

	// check cross
	xWins = (b[0][0] > 0 && b[1][1] > 0 && b[2][2] > 0) || (b[2][0] > 0 && b[1][1] > 0 && b[0][2] > 0) || xWins
	oWins = (b[0][0] < 0 && b[1][1] < 0 && b[2][2] < 0) || (b[2][0] < 0 && b[1][1] < 0 && b[0][2] < 0) || oWins

out:
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			movesLeft = b[x][y] == 0
			if movesLeft {
				break out
			}
		}
	}

	if xWins {
		return true, 1
	} else if oWins {
		return true, -1
	} else if !movesLeft {
		return true, 0
	}

	return false, 0
}

func (g *Game) End() {
	g.Over = true
	g.Player1.EndGame(false)
	g.Player2.EndGame(false)
}

type Player struct {
	GameId      int
	CurrentTurn bool
	in          <-chan *Message
	out         chan<- *Message
	done        <-chan bool
	err         <-chan error
	disconnect  chan<- int
}

func (p *Player) Negotiating() {
	p.out <- &Message{0, "Server", "Negotiating", 0, 0, false}
}

func (p *Player) EndGame(win bool) {
	if win {
		p.out <- &Message{3, "Server", "You Win!", 0, 0, false}
	} else {
		p.out <- &Message{3, "Server", "You Fail!", 0, 0, false}
	}

	p.disconnect <- 1000
}

func (p *Player) StartTurn() {
	p.CurrentTurn = true
	p.out <- &Message{1, "Server", "Your Turn", 0, 0, false}
}

func (p *Player) EndTurn() {
	p.CurrentTurn = false
	p.out <- &Message{2, "Server", "Opponent's Turn", 0, 0, false}
}

func (p *Player) SendMove(x, y, move int, sender string, isX bool) {
	p.out <- &Message{-5, sender, fmt.Sprintf("played (%d, %d)", x, y), x, y, isX}
}

func (p *Player) UpdateStatus(message string) {
	p.out <- &Message{-1, "Server", message, 0, 0, false}
}

type Message struct {
	Type    int    `json:"type"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	IsX     bool   `json:"isX"`
}
