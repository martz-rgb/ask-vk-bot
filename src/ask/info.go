package ask

import (
	"database/sql"
)

type Info struct {
	VkID     int            `db:"vk_id"`
	Gallery  sql.NullString `db:"gallery"`
	Birthday sql.NullTime   `db:"birthday"`
}
