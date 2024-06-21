package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"de.whatwapp/app/connection"
	m "de.whatwapp/app/model"
	"de.whatwapp/app/model/response"
	"de.whatwapp/app/store"
	"github.com/davecgh/go-spew/spew"
)

type MatchController Controller[m.Match]

const timeoutInSeconds = 30

func CreateMatchApi(c *Controller[m.Match], playerController *Controller[m.Player], tableController *Controller[m.Table]) {
	fmt.Println()
	matchController := MatchController(*c)
	root := fmt.Sprintf("/%s", c.model)

	c.DeleteMultiple(fmt.Sprintf("%s/multiple", root), nil)
	(&matchController).findMatch(fmt.Sprintf("%s/find", root), playerController, tableController)

	c.updateDB()
}

// Matchmaking godoc
//
//	@Summary		Matchmaking
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Failure		400	{object}	interface{}
//	@Failure		404	{object}	interface{}
//	@Failure		500	{object}	interface{}
//	@Router			/matches/find [get]
func (c *MatchController) findMatch(path string, playerController *Controller[m.Player], tableController *Controller[m.Table]) {
	fmt.Println("GET", path)
	c.router.Get(path, func(w http.ResponseWriter, r *http.Request) {
		socket, err := connection.NewWebSocket(w, r)
		if err != nil {
			HandleResponse(w, nil, err)
			return
		}
		table, err := getTable(r, tableController)
		if err != nil {
			socket.Write(response.Response{Message: err.Error()})
			socket.Close()
			return
		}
		playerMatchmaking, err := c.getPlayerMatchmaking(r, playerController, table)
		if err != nil {
			socket.Write(response.Response{Message: err.Error()})
			socket.Close()
			return
		}
		c.startMatchmaking(
			r,
			playerMatchmaking,
			table,
			time.Duration(timeoutInSeconds*float64(time.Second)),
			socket,
		)
	})
}

func (c *MatchController) getPlayerMatchmaking(r *http.Request, playerController *Controller[m.Player], table *m.Table) (*m.Match, error) {
	player, err := getPlayer(r, playerController)
	if err != nil {
		return nil, err
	}
	if err := c.isSearchingForAMatch(player); err != nil {
		return nil, err
	}
	playerMatch, err := c.addPlayerToMatchWaitingList(player, table)
	if err != nil {
		return nil, err
	}
	return playerMatch, nil
}

func getTable(r *http.Request, tableController *Controller[m.Table]) (*m.Table, error) {
	tableName, err := strconv.Atoi(r.URL.Query().Get("table"))
	if err != nil {
		return nil, err
	}
	table, err := tableController.Store.FindOne(&m.Table{
		Name: tableName,
	}, nil)
	return table, err
}

func (c *MatchController) startMatchmaking(r *http.Request, playerMatchmaking *m.Match, table *m.Table, timeout time.Duration, socket *connection.WebSocket) {

	done := make(chan bool)
	quit := make(chan bool)
	clientInput := make(chan interface{})

	socket.Write(response.Response{Message: "starting matchmaking"})
	socket.Listen(clientInput)

	go c.findAMatch(quit, done, playerMatchmaking, table, socket)
	select {
	//see this as user cancel
	case <-clientInput:
		err := c.deleteMatch(playerMatchmaking)
		if err != nil {
			socket.Write(err)
		}
		quit <- true
	case <-done:
		shouldStopMathmaking := StopMatchmakingIfPossible(c, playerMatchmaking, table, socket)
		if shouldStopMathmaking {
			return
		}
		c.startMatchmaking(r, playerMatchmaking, table, timeout, socket)
	case <-time.After(timeout):
		shouldStopMathmaking := StopMatchmakingIfPossible(c, playerMatchmaking, table, socket)
		if shouldStopMathmaking {
			return
		}
		err := c.deleteMatch(playerMatchmaking)
		if err != nil {
			socket.Write(response.Response{Message: err.Error()})
		} else {
			socket.Write(response.Response{Message: "matchmaking timed out"})
		}
		socket.Close()
		quit <- true
	}
}

func StopMatchmakingIfPossible(c *MatchController, playerMatchmaking *m.Match, table *m.Table, socket *connection.WebSocket) bool {
	if match, err := c.Store.GetPlayerMatch(playerMatchmaking); len(*match) == 1 && err == nil {
		err := c.Store.CompleteMatchIfPossible(playerMatchmaking, table)
		if err != nil {
			socket.Write(response.Response{Message: err.Error()})
		} else {
			socket.Write(response.Response{Message: (*match)[0].Server})
		}
		socket.Close()
		return true
	}
	return false
}

func (c *MatchController) deleteMatch(match *m.Match) error {
	_, err := c.Store.Delete(match)
	return err
}

func (c *MatchController) findAMatch(quit chan bool, done chan bool, playerMatch *m.Match, table *m.Table, socket *connection.WebSocket) {
	for {
		select {
		case <-quit:
			return
		default:
			socket.Write(response.Response{Message: "searching match"})
			if match, err := c.Store.GetPlayerMatch(playerMatch); len(*match) == 1 && err == nil {
				fmt.Println("I am in match. Joining")
				done <- true
				return
			}
			playersForMatch, err := c.Store.FindPlayersForMatch(playerMatch)
			fmt.Printf("player: %d\n", playerMatch.PlayerId)
			fmt.Println("players for match")
			spew.Dump(playersForMatch)
			if err != nil {
				fmt.Println(err.Error())
				done <- false
				return
			}
			if len(*playersForMatch) > 0 {
				matchStore := store.MatchStore(*c.Store)
				err := matchStore.SetGameForMatch(playerMatch, table)
				if err == nil {
					done <- true
					return
				}
				fmt.Println(err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *MatchController) isSearchingForAMatch(player *m.Player) error {
	matchFound, err := c.Store.FindOne(&m.Match{
		PlayerId: player.ID,
		Server:   nil,
	}, nil)
	if err == nil && matchFound != nil {
		return fmt.Errorf("player %s is already SEARCHING FOR or PLAYING a match", player.Username)
	}
	return nil
}

func (c *MatchController) addPlayerToMatchWaitingList(player *m.Player, table *m.Table) (*m.Match, error) {
	if player == nil {
		return nil, fmt.Errorf("missing player for match")
	}
	if table == nil {
		return nil, fmt.Errorf("missing table for match")
	}
	match, err := c.Store.AddOne(&m.Match{
		PlayerId:  player.ID,
		League:    &player.League,
		TableName: table.Name,
	})
	if err != nil {
		return nil, err
	}
	return match, nil
}

func getPlayer(r *http.Request, playerController *Controller[m.Player]) (*m.Player, error) {
	username := r.URL.Query().Get("username")
	if username == "" {
		return nil, fmt.Errorf("username missing")
	}
	return playerController.Store.FindOne(&m.Player{
		Username: r.URL.Query().Get("username"),
	}, nil)
}
