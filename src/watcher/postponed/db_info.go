package postponed

import "ask-bot/src/ask"

type DBInfo struct {
	polls []ask.PendingPoll
	// greetings
	// leavings
}

func NewDBInfo(c *Controls) (*DBInfo, error) {
	polls, err := c.Ask.PendingPolls()
	if err != nil {
		return nil, err
	}

	return &DBInfo{
		polls: polls,
	}, nil
}
