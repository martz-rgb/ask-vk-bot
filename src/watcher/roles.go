package watcher

import "fmt"

func (c *Controls) CheckAlbums() error {
	roles, err := c.Ask.Roles()
	if err != nil {
		return err
	}

	albums := make(map[string]int)

	var id int
	var vk_err error
	for _, role := range roles {
		if !role.Album.Valid {
			id, vk_err = c.Admin.CreateAlbum(role.CaptionName)
			if vk_err != nil {
				break
			}

			albums[role.Name] = id
		}
	}

	if len(albums) > 0 {
		err = c.Ask.ChangeAlbums(albums)
		if err != nil {
			return err
		}
	}

	return vk_err
}

func (c *Controls) CheckBoards() error {
	roles, err := c.Ask.Roles()
	if err != nil {
		return err
	}

	boards := make(map[string]int)

	var id int
	var vk_err error
	for _, role := range roles {
		if !role.Board.Valid {
			text := fmt.Sprintf("Здесь можно задать вопрос %s!", role.AccusativeName)
			id, vk_err = c.Admin.CreateBoard(role.CaptionName, text, "")
			if vk_err != nil {
				break
			}

			boards[role.Name] = id
		}
	}

	if len(boards) > 0 {
		err = c.Ask.ChangeBoards(boards)
		if err != nil {
			return err
		}
	}

	return vk_err
}
