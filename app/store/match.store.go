package store

import (
	"fmt"

	m "de.whatwapp/app/model"
	"gorm.io/gorm"
)

type MatchStore Store[m.Match]

func (store *Store[T]) FindMatch(match *m.Match) (*[]m.Match, error) {
	var matches []m.Match
	minLeague, maxLeague, table := getMatchParmeters(match)
	result := store.DB.Debug().Table(store.TableName).Where(
		"league > ? AND league < ? AND table_name = ? AND server IS NULL", minLeague, maxLeague, table,
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

func (matchStore *MatchStore) SetGameForMatch(match *m.Match, table *m.Table) (*[]m.Match, error) {
	store := Store[m.Match](*matchStore)
	var matches *[]m.Match
	err := matchStore.DB.Transaction(func(tx *gorm.DB) error {
		fmt.Println("creating match")
		var matches []m.Match
		minLeague, maxLeague, tableName := getMatchParmeters(match)
		result := tx.Debug().Table(store.TableName).Where(
			"league > ? AND league < ? AND table_name = ? AND server IS NULL", minLeague, maxLeague, tableName,
		).Order(
			"created_at ASC",
		).Find(&matches)

		if result.Error != nil {
			fmt.Println("error at search matches")
			fmt.Println(result.Error.Error())
			return result.Error
		}
		if len(matches) < *table.Min {
			return fmt.Errorf("not enough players")
		}
		if len(matches) > *table.Max {
			tmp := matches[:*table.Max]
			matches = tmp
		}
		var server *m.Server
		tx.Debug().Table(m.GAME_MODEL).Order("created_at asc").First(&server)
		if server == nil {
			fmt.Println("error at find serer")
			return fmt.Errorf("no server available")
		}
		for i := range matches {
			matches[i].Server = &server.ID
		}
		result = tx.Debug().Table(store.TableName).Save(&matches)
		if result.Error != nil {
			fmt.Println(result.Error.Error())
			return result.Error
		}
		return nil
	})
	return matches, err
}
