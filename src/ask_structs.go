package main

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"
)

type Administration struct {
	VkID int `db:"vk_id"`
}

// TO-DO: time.Time is dangerous but i want to try
type Person struct {
	VkID     int            `db:"vk_id"`
	Name     sql.NullString `db:"name"`
	Gallery  sql.NullString `db:"gallery"`
	Birthday sql.NullTime   `db:"birthday"`
}

type Role struct {
	Name        string         `db:"name"`
	Tag         string         `db:"tag"`
	ShownName   string         `db:"shown_name"`
	CaptionName sql.NullString `db:"caption_name"`
	Album       sql.NullString `db:"album_link"`
	Board       sql.NullString `db:"board_link"`
}

type Points struct {
	Person    int       `db:"person"`
	Diff      int       `db:"diff"`
	Cause     string    `db:"cause"`
	Timestamp time.Time `db:"timestamp"`
}

type MemberStatus string

var MemberStatuses = struct {
	Active MemberStatus
	Freeze MemberStatus
}{
	Active: "Active",
	Freeze: "Freeze",
}

func (s MemberStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *MemberStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("MemberStatus is not nullable")
	}
	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			// check if is valid
			if v != string(MemberStatuses.Active) && v != string(MemberStatuses.Freeze) {
				return errors.New("value is not valid MemberStatus value")
			}
			*s = MemberStatus(v)
			return nil
		}
	}
	return errors.New("failed to scan MemberStatus")
}

type Member struct {
	Id     int          `db:"id"`
	Person int          `db:"person"`
	Role   string       `db:"role"`
	Status MemberStatus `db:"status"`
}

type DeadlineCause string

var DeadlineCauses = struct {
	Init   DeadlineCause
	Answer DeadlineCause
	Delay  DeadlineCause
	Rest   DeadlineCause
	Freeze DeadlineCause
	Other  DeadlineCause
}{
	Init:   "Init",
	Answer: "Answer",
	Delay:  "Delay",
	Rest:   "Rest",
	Freeze: "Freeze",
	Other:  "Other",
}

func (c DeadlineCause) Value() (driver.Value, error) {
	return string(c), nil
}

func (c *DeadlineCause) Scan(value interface{}) error {
	if value == nil {
		return errors.New("DeadlineCause is not nullable")
	}
	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			// check if is valid
			if v != string(DeadlineCauses.Init) &&
				v != string(DeadlineCauses.Answer) &&
				v != string(DeadlineCauses.Delay) &&
				v != string(DeadlineCauses.Rest) &&
				v != string(DeadlineCauses.Freeze) &&
				v != string(DeadlineCauses.Other) {
				return errors.New("value is not valid DeadlineCause value")
			}
			*c = DeadlineCause(v)
			return nil
		}
	}
	return errors.New("failed to scan DeadlineCause")
}

type Deadline struct {
	Member    int           `db:"member"`
	Diff      int           `db:"diff"` // unix time in seconds!
	Type      DeadlineCause `db:"type"`
	Cause     string        `db:"cause"`
	Timestamp time.Time     `db:"timestamp"`
}
