package store

import (
	"fmt"
	"time"

	m "de.whatwapp/app/model"
	"gorm.io/gorm"
)

type MatchStore Store[m.Match]

func (store *Store[T]) FindPlayersForMatch(match *m.Match) (*[]m.Match, error) {
	var matches []m.Match
	minLeague, maxLeague, tableName := getMatchParmeters(match)
	result := store.DB.Debug().Table(store.TableName).Where(
		"league > ? AND league < ? AND table_name = ? AND deleted_at IS NULL AND playing = false", minLeague, maxLeague, tableName,
	).Order(
		"created_at ASC",
	).Find(&matches)
	return &matches, result.Error
}

func (store *Store[T]) GetPlayerMatch(match *m.Match) (*[]m.Match, error) {
	var matches []m.Match
	result := store.DB.Debug().Table(store.TableName).Where(
		"player_id = ? AND deleted_at IS NULL AND server IS NOT NULL", match.PlayerId,
	).Find(&matches)
	return &matches, result.Error
}

func (store *Store[T]) CompleteMatchIfPossible(match *m.Match, table *m.Table) error {
	playersForMatch, err := store.FindPlayersForMatch(match)
	if err != nil {
		return err
	}
	if len(*playersForMatch) == *table.MaxPlayers {
		fmt.Println("max players reached")
		return store.startGame(playersForMatch)
	}
	firstJoin := time.Now()
	for _, p := range *playersForMatch {
		if p.UpdatedAt.Before(firstJoin) {
			firstJoin = p.UpdatedAt
		}
	}
	if firstJoin.Add(time.Duration(*table.MaxWaitTime) * time.Second).Before(time.Now()) {
		fmt.Println("game should start")
		return store.startGame(playersForMatch)
	}
	return nil
}

func (store *Store[T]) startGame(playersForMatch *[]m.Match) error {
	for i := range *playersForMatch {
		(*playersForMatch)[i].Playing = true
	}
	result := store.DB.Table(store.TableName).Save(*playersForMatch)
	return result.Error
}

func getMatchParmeters(match *m.Match) (int, int, int) {
	league := *match.League
	minLeague := 0
	maxLeague := 100
	if league-1 > 0 {
		minLeague = league - 1
	}
	if league+2 < 100 {
		maxLeague = league + 2
	}
	return minLeague, maxLeague, match.TableName
}

func (matchStore *MatchStore) SetGameForMatch(match *m.Match, table *m.Table) error {
	store := Store[m.Match](*matchStore)
	err := matchStore.DB.Transaction(func(tx *gorm.DB) error {
		fmt.Println("creating match")
		minLeague, maxLeague, tableName := getMatchParmeters(match)
		//get players that are waiting to join a game
		var availablePlayers []m.Match
		result := tx.Debug().Table(store.TableName).Where(
			"league > ? AND league < ? AND table_name = ? AND server IS NULL", minLeague, maxLeague, tableName,
		).Order(
			"created_at ASC",
		).Find(&availablePlayers)
		if result.Error != nil {
			fmt.Println(result.Error.Error())
			return result.Error
		}

		//get players that are in a lobby for a game (logged in but game not started yet)
		var waitingPlayers []m.Match
		result = tx.Debug().Table(store.TableName).Where(
			"league > ? AND league < ? AND table_name = ? AND server IS NOT NULL", minLeague, maxLeague, tableName,
		).Find(&waitingPlayers)
		if result.Error != nil {
			fmt.Println(result.Error.Error())
			return result.Error
		}
		// not enoght players are available to start a game
		if len(waitingPlayers) == 0 && len(availablePlayers) < *table.MinPlayers {
			return fmt.Errorf("not enough players")
		}

		// get the players that can join a lobby
		numPlayersCanJoin := *table.MaxPlayers - len(waitingPlayers)
		if len(availablePlayers) > numPlayersCanJoin {
			availablePlayers = availablePlayers[:numPlayersCanJoin]
		}

		//get the server where the player can connect to
		// either is a new one (min player number reached)
		// or the player is joining layers in a lobby
		gameServer := m.Server{}
		if len(waitingPlayers) > 0 {
			gameServer.ID = *(waitingPlayers[0].Server)
		} else {
			//get server from pool of available servers and delete it from it after that
			result := tx.Debug().Table(m.SERVER_MODEL).Order("created_at asc").First(&gameServer)
			if result.Error != nil {
				return result.Error
			}
			result = tx.Debug().Table(m.SERVER_MODEL).Delete(&gameServer, &gameServer)
			if result.Error != nil {
				return result.Error
			}
		}
		// set the server to which the players will have to join
		for i := range availablePlayers {
			availablePlayers[i].Server = &gameServer.ID
		}
		result = tx.Debug().Table(store.TableName).Save(&availablePlayers)
		if result.Error != nil {
			fmt.Println(result.Error.Error())
			return result.Error
		}
		return nil
	})
	return err
}
