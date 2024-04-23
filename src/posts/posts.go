package posts

import (
	"ask-bot/src/ask"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/object"
)

type Kind int

var Kinds = struct {
	Unknown    Kind
	Invalid    Kind
	Poll       Kind
	Acceptance Kind
	Answer     Kind
	FreeAnswer Kind
	Leaving    Kind
}{
	Unknown:    0,
	Invalid:    1,
	Poll:       2,
	Acceptance: 3,
	Answer:     4,
	FreeAnswer: 5,
	Leaving:    6,
}

type Poll struct {
	ID      int
	Closed  bool
	Answers []object.PollsAnswer
}

type Post struct {
	Kind  Kind
	Roles []ask.Role

	ID   int
	Date time.Time

	Poll *Poll
}

func Parse(vk_post *object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) *Post {
	post := &Post{
		ID:   vk_post.ID,
		Date: time.Unix(int64(vk_post.Date), 0),
	}

	post.complete(vk_post, dictionary, organization)

	return post
}

func ParseMany(vk_posts []object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) []Post {
	posts := make([]Post, len(vk_posts))

	for i, vk_post := range vk_posts {
		posts[i].ID = vk_post.ID
		posts[i].Date = time.Unix(int64(vk_post.Date), 0)

		posts[i].complete(&vk_post, dictionary, organization)
	}

	return posts
}

func (p *Post) complete(vk_post *object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) {
	tags := regexp.MustCompile(`#([\w@]+)`).FindAllString(vk_post.Text, -1)
	p.Roles = FindRoles(tags, dictionary)

	var kind Kind
	if len(p.Roles) > 0 {
		kind = Kinds.Answer
	} else {
		kind = Kinds.Unknown
	}

	count := 0

	if slices.Contains(tags, organization.PollHashtag) {
		kind = Kinds.Poll
		count++

		if len(p.Roles) != 1 {
			kind = Kinds.Invalid
		}

		var poll *Poll
		for _, attachment := range vk_post.Attachments {
			if attachment.Type != object.AttachmentTypePoll {
				continue
			}

			poll = &Poll{
				ID:      attachment.Poll.ID,
				Closed:  bool(attachment.Poll.Closed),
				Answers: attachment.Poll.Answers,
			}
		}

		if poll == nil {
			kind = Kinds.Invalid
		} else {
			p.Poll = poll
		}
	}

	if slices.Contains(tags, organization.AcceptanceHashtag) {
		kind = Kinds.Acceptance
		count++
	}
	if slices.Contains(tags, organization.FreeAnswerHashtag) {
		kind = Kinds.FreeAnswer
		count++
	}
	if slices.Contains(tags, organization.LeavingHashtag) {
		kind = Kinds.Leaving
		count++
	}

	if count > 1 {
		kind = Kinds.Invalid
	}

	p.Kind = kind
}

func FindRoles(tags []string, dictionary []ask.Role) []ask.Role {
	var found []ask.Role

	for _, t := range tags {
		index, ok := slices.BinarySearchFunc(
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

func ToTime(posts []Post) []time.Time {
	var result []time.Time
	for _, p := range posts {
		result = append(result, p.Date)
	}

	slices.SortFunc(result, func(a, b time.Time) int {
		if a.After(b) {
			return 1
		} else if a.Before(b) {
			return -1
		}

		return 0
	})

	return result
}
