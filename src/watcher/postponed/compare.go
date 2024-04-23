package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
	"slices"
	"strings"
)

func exclude(actual *VKInfo, desired *DBInfo) (new *DBInfo, invalid []posts.Post) {
	// polls, acceptances & leavings

	polls, old := exclude_polls(actual.posts[posts.Kinds.Poll], desired.polls)
	invalid = append(invalid, old...)

	return &DBInfo{
		polls: polls,
	}, invalid
}

func exclude_polls(actual []posts.Post, desired []ask.PendingPoll) (new []ask.PendingPoll, invalid []posts.Post) {
	// assumption that desired is sorted by role?

	for i := range actual {
		index, ok := slices.BinarySearchFunc(desired, actual[i], func(pp ask.PendingPoll, p posts.Post) int {
			return strings.Compare(pp.Name, p.Roles[0].Name)
		})

		if !ok {
			invalid = append(invalid, actual[i])
			continue
		}

		// check for correctness
		// count of members
		// TO-DO should be with ask configuration
		if len(actual[i].Poll.Answers) != desired[index].Count+1 {
			invalid = append(invalid, actual[i])
			continue
		}

		desired = append(desired[:index], desired[index+1:]...)
	}

	return desired, invalid
}

// func exclude_acceptances(db *DBInfo, vk *VKInfo) []posts.Post {
// 	return nil
// }

// func exclude_leavings(db *DBInfo, vk *VKInfo) []posts.Post {
// 	return nil
// }
