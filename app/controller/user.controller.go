package controller

import (
	"fmt"
	"net/http"

	"de.whatwapp/app/auth"
	mw "de.whatwapp/app/middleware"
	m "de.whatwapp/app/model"
)

func CreateUserApi(c *Controller[m.User]) {
	fmt.Println("")
	root := fmt.Sprintf("/%s", c.model)

	c.Create(root)
	c.Update(root)
	c.GetToken(c, root+"/token")

	c.updateDB()
}

// User GetToken godoc
//
//	@Summary		User GetToken
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Failure		400	{object}	interface{}
//	@Failure		404	{object}	interface{}
//	@Failure		500	{object}	interface{}
//	@Router			/user/token [get]
func (c *Controller[T]) GetToken(controller *Controller[m.User], path string, handlers ...mw.Middleware) {
	fmt.Println("GET", path)
	controller.router.
		With(mw.GetMiddlewares(handlers)...).
		Get(path, func(w http.ResponseWriter, r *http.Request) {
			data, err := m.CreateObjectFromBody[m.User](r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if data.Email == nil || data.Password == nil {
				HandleResponse(w, nil, fmt.Errorf("both email and password must be specified in role request"))
				return
			}
			response, err := controller.Store.FindOne(data, nil)
			if response != nil {
				token, err := auth.CreateToken(response.Role)
				HandleResponse(w, token, err)
				return
			}
			HandleResponse(w, response, err)
		})
}
