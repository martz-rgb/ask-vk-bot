package posts

import (
	"ask-bot/src/ask"
	"regexp"
	"slices"
	"strings"

	"github.com/SevereCloud/vksdk/v2/object"
)

type Kind int

const (
	Unknown    Kind = 0
	Invalid    Kind = 1
	Poll       Kind = 2
	Answer     Kind = 3
	FreeAnswer Kind = 4
	Leaving    Kind = 5
)

type Post struct {
	Tags  []string
	Kind  Kind
	Roles []ask.Role

	Vk *object.WallWallpost
}

func Parse(vk_post *object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) *Post {
	post := &Post{
		Tags: regexp.MustCompile(`#([\w@]+)`).FindAllString(vk_post.Text, -1),
		Vk:   vk_post,
	}

	post.complete(dictionary, organization)

	return post
}

func ParseMany(vk_posts []object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) []Post {
	posts := make([]Post, len(vk_posts))

	for i, vk_post := range vk_posts {
		posts[i].Tags = regexp.MustCompile(`#([\w@]+)`).FindAllString(vk_post.Text, -1)
		posts[i].Vk = &vk_posts[i]

		posts[i].complete(dictionary, organization)
	}

	return posts
}

func (p *Post) complete(dictionary []ask.Role, organization *ask.OrganizationHashtags) {
	p.Roles = FindRoles(p.Tags, dictionary)

	var kind Kind
	if len(p.Roles) > 0 {
		kind = Answer
	} else {
		kind = Unknown
	}

	count := 0

	if slices.Contains[[]string](p.Tags, organization.PollHashtag) {
		kind = Poll
		count++

		if len(p.Roles) != 1 {
			kind = Invalid
		}
	}
	if slices.Contains[[]string](p.Tags, organization.FreeAnswerHashtag) {
		kind = FreeAnswer
		count++
	}
	if slices.Contains[[]string](p.Tags, organization.LeavingHashtag) {
		kind = Leaving
		count++
	}

	if count > 1 {
		kind = Invalid
	}

	p.Kind = kind
}

func FindRoles(tags []string, dictionary []ask.Role) []ask.Role {
	var found []ask.Role

	for _, t := range tags {
		index, ok := slices.BinarySearchFunc[[]ask.Role](
			dictionary,
			t,
			func(r ask.Role, s string) int {
				return strings.Compare(r.Hashtag, s)
			})

		if ok {
			found = append(found, dictionary[index])
		}
	}

	return found
}
