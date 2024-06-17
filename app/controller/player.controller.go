package controller

import (
	"fmt"

	"de.whatwapp/app/model"
)

func CreatePlayerApi(c *Controller[model.Player]) {
	fmt.Println("")
	root := fmt.Sprintf("/%s", c.model)

	c.Create(root)
	c.Update(root)
	c.Delete(root, nil)

	c.updateDB()
}
