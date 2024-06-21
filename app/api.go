package app

import (
	"fmt"
	"log"
	"net/http"

	c "de.whatwapp/app/controller"
	"de.whatwapp/app/db"
	m "de.whatwapp/app/model"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	_ "github.com/lib/pq"
)

func NewApp() {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	db, err := db.ConnectToDB(m.MysqlConfig)
	if err != nil {
		panic(err)
	}

	user := c.NewController[m.User](db, router, m.USER_MODEL)
	match := c.NewController[m.Match](db, router, m.MATCH_MODEL)
	player := c.NewController[m.Player](db, router, m.PLAYER_MODEL)
	table := c.NewController[m.Table](db, router, m.TABLE_MODEL)
	servers := c.NewController[m.Server](db, router, m.SERVER_MODEL)

	c.CreateServerApi(servers)
	c.CreateTableApi(table)
	c.CreatePlayerApi(player)
	c.CreateMatchApi(match, player, table)
	c.CreateUserApi(user)

	conf := m.GetServerConf()
	fmt.Println()
	fmt.Println("API created")
	fmt.Printf("STARTING SERVER on port %d", conf.Port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
}
