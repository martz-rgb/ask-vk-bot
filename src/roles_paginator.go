package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type RolesPaginator struct {
	roles       []Role
	page        int
	total_pages int

	rows int
	cols int
}

var RowsCount int = 2
var ColsCount int = 3

func NewRolesPaginator(roles []Role, rows int, cols int) *RolesPaginator {
	return &RolesPaginator{
		roles: roles,
		page:  0,
		// ceil function
		total_pages: 1 + (len(roles)-1)/(rows*cols),
		rows:        rows,
		cols:        cols,
	}
}

func (k *RolesPaginator) Next() {
	k.page += 1
	if k.page >= k.total_pages {
		k.page = k.total_pages - 1
	}
}

func (k *RolesPaginator) Previous() {
	k.page -= 1
	if k.page < 0 {
		k.page = 0
	}
}

func (k *RolesPaginator) ChangeRoles(roles []Role) {
	k.roles = roles
	k.page = 0
	k.total_pages = 1 + (len(roles)-1)/(k.rows*k.cols)
}

func (k *RolesPaginator) Role(name string) (*Role, error) {
	for _, role := range k.roles {
		if role.Name == name {
			return &role, nil
		}
	}

	err := errors.New("failed to find role in list")
	return nil, zaperr.Wrap(err, "",
		zap.String("role", name),
		zap.Any("roles", k.roles))
}

func (k *RolesPaginator) Buttons() [][]Button {
	buttons := [][]Button{}

	for i := 0; i < k.rows; i++ {
		if i*k.cols >= len(k.roles) {
			break
		}

		buttons = append(buttons, []Button{})

		for j := 0; j < k.cols; j++ {
			index := i*k.cols + j + k.page*(k.rows*k.cols)

			if index >= len(k.roles) {
				i = k.rows
				break
			}

			// maybe send index?..
			buttons[i] = append(buttons[i], Button{
				Label: k.roles[index].ShownName,
				Color: "secondary",

				Command: "roles",
				Value:   k.roles[index].Name,
			})
		}
	}

	// + доп ряд с функциональными кнопками
	controls := []Button{}

	if k.page > 0 {
		controls = append(controls, Button{
			Label: "<<",
			Color: "primary",

			Command: "previous",
		})
	}

	if k.page < k.total_pages-1 {
		controls = append(controls, Button{
			Label: ">>",
			Color: "primary",

			Command: "next",
		})
	}

	controls = append(controls, Button{
		Label: "Назад",
		Color: "negative",

		Command: "back",
	})

	buttons = append(buttons, controls)

	return buttons
}
