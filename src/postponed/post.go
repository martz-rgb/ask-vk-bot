package postponed

import (
	"ask-bot/src/ask"
	"regexp"
	"slices"
	"strings"

	"github.com/SevereCloud/vksdk/v2/object"
)

type PostKind int

const (
	Unknown    PostKind = 0
	Invalid    PostKind = 1
	Poll       PostKind = 2
	FreeAnswer PostKind = 3
	Leaving    PostKind = 4
)

type Post struct {
	Tags  []string
	Kind  PostKind
	Roles []ask.Role

	Vk *object.WallWallpost
}

func FindRoles(tags []string, roles []ask.Role) []ask.Role {
	var found []ask.Role

	for _, t := range tags {
		index, ok := slices.BinarySearchFunc[[]ask.Role](
			roles,
			t,
			func(r ask.Role, s string) int {
				return strings.Compare(r.Hashtag, s)
			})

		if ok {
			found = append(found, roles[index])
		}
	}

	return found
}

func Parse(roles_dictionary []ask.Role, organization_tags *ask.OrganizationHashtags, vk_posts []object.WallWallpost) ([]Post, error) {
	var posts []Post

	for _, vk_post := range vk_posts {
		tags := regexp.MustCompile(`#([\w@]+)`).FindAllString(vk_post.Text, -1)
		kind := Kind(tags, organization_tags)
		roles := FindRoles(tags, roles_dictionary)

		posts = append(posts,
			Post{
				tags,
				kind,
				roles,
				&vk_post,
			})
	}

	return posts, nil
}

func Kind(tags []string, organization *ask.OrganizationHashtags) PostKind {
	poll := slices.Contains[[]string](tags, organization.PollHashtag)
	free_answer := slices.Contains[[]string](tags, organization.FreeAnswerHashtag)
	leaving := slices.Contains[[]string](tags, organization.LeavingHashtag)

	kind := Unknown
	count := 0

	if poll {
		kind = Poll
		count++
	}
	if free_answer {
		kind = FreeAnswer
		count++
	}
	if leaving {
		kind = Leaving
		count++
	}

	if count > 1 {
		return Invalid
	}
	return kind
}
