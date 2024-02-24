package main

import "ask-bot/src/ask"

type User struct {
	id int

	members []ask.Member
	roles   []ask.Role
	filled  bool
}

func (u *User) fill(a *ask.Ask) error {
	members, err := a.MembersByVkID(u.id)
	if err != nil {
		return err
	}

	u.members = members
	u.roles = make([]ask.Role, len(u.members))

	for i := range u.members {
		role, err := a.Role(u.members[i].Role)
		if err != nil {
			return err
		}

		u.roles[i] = role
	}

	u.filled = true
	return nil
}

func (u *User) Members(a *ask.Ask) ([]ask.Member, error) {
	if u.filled {
		return u.members, nil
	}

	err := u.fill(a)
	if err != nil {
		return nil, err
	}

	return u.members, nil
}

func (u *User) Roles(a *ask.Ask) ([]ask.Role, error) {
	if u.filled {
		return u.roles, nil
	}

	err := u.fill(a)
	if err != nil {
		return nil, err
	}

	return u.roles, nil
}

// members and roles are filled respectively
func (u *User) MembersRoles(a *ask.Ask) ([]ask.Member, []ask.Role, error) {
	if u.filled {
		return u.members, u.roles, nil
	}

	err := u.fill(a)
	if err != nil {
		return nil, nil, err
	}

	return u.members, u.roles, nil
}
