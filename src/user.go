package main

type User struct {
	id int

	members []Member
	roles   []Role
	filled  bool
}

func (u *User) fill(ask *Ask) error {
	members, err := ask.MembersById(u.id)
	if err != nil {
		return err
	}

	u.members = members
	u.roles = make([]Role, len(u.members))

	for i := range u.members {
		role, err := ask.Role(u.members[i].Role)
		if err != nil {
			return err
		}

		u.roles[i] = role
	}

	u.filled = true
	return nil
}

func (u *User) Members(ask *Ask) ([]Member, error) {
	if u.filled {
		return u.members, nil
	}

	err := u.fill(ask)
	if err != nil {
		return nil, err
	}

	return u.members, nil
}

func (u *User) Roles(ask *Ask) ([]Role, error) {
	if u.filled {
		return u.roles, nil
	}

	err := u.fill(ask)
	if err != nil {
		return nil, err
	}

	return u.roles, nil
}

// members and roles are filled respectively
func (u *User) MembersRoles(ask *Ask) ([]Member, []Role, error) {
	if u.filled {
		return u.members, u.roles, nil
	}

	err := u.fill(ask)
	if err != nil {
		return nil, nil, err
	}

	return u.members, u.roles, nil
}
