package controller

import (
	"fmt"

	m "de.whatwapp/app/model"
)

func CreateTableApi(c *Controller[m.Table]) {
	root := fmt.Sprintf("/%s", c.model)

	c.Create(root)
	c.Update(root)
	c.Delete(root, nil)

	c.updateDB()
}
