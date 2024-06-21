package controller

import (
	"fmt"

	m "de.whatwapp/app/model"
)

func CreateServerApi(c *Controller[m.Server]) {
	fmt.Println()
	root := fmt.Sprintf("/%s", c.model)

	c.Create(root)
	c.Delete(root, nil)

	c.updateDB()
}
