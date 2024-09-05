package posts

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/schedule"
	"ask-bot/src/vk"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/object"
)

// binary mask
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
	Unknown:    0b1,
	Invalid:    0b10,
	Poll:       0b100,
	Acceptance: 0b1000,
	Answer:     0b10000,
	FreeAnswer: 0b100000,
	Leaving:    0b1000000,
}

func ParseKinds(mask Kind) []Kind {
	var kinds []Kind

	// ATTENTION
	index := 0b1000000

	for index > 0 {
		if int(mask)&index == 1 {
			kinds = append(kinds, Kind(index))
		}

		index >>= 1
	}

	return kinds
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

	//Poll *Poll
}

func Parse(vk_post *object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) *Post {
	post := &Post{
		ID:   vk_post.ID,
		Date: time.Unix(int64(vk_post.Date), 0),
	}

	post.complete(vk_post.Text, dictionary, organization)

	return post
}

func ParseFromParams(id int, params vk.PostParams, dictionary []ask.Role, organization *ask.OrganizationHashtags) *Post {
	post := &Post{
		Kind:  Kinds.Unknown,
		Roles: nil,

		ID:   id,
		Date: params.PublishDate,

		//Poll: nil,
	}

	// find kind & roles
	post.complete(params.Text, dictionary, organization)

	return post
}

func (p *Post) complete(text string, dictionary []ask.Role, organization *ask.OrganizationHashtags) {
	tags := regexp.MustCompile(`#([\w@]+)`).FindAllString(text, -1)
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

		// var poll *Poll
		// for _, attachment := range vk_post.Attachments {
		// 	if attachment.Type != object.AttachmentTypePoll {
		// 		continue
		// 	}

		// 	poll = &Poll{
		// 		ID:      attachment.Poll.ID,
		// 		Closed:  bool(attachment.Poll.Closed),
		// 		Answers: attachment.Poll.Answers,
		// 	}
		// }

		// if poll == nil {
		// 	kind = Kinds.Invalid
		// } else {
		// 	p.Poll = poll
		// }
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

type Posts map[Kind][]Post

func ParseMany(vk_posts []object.WallWallpost, dictionary []ask.Role, organization *ask.OrganizationHashtags) Posts {
	posts := make(Posts, len(vk_posts))

	for i := range vk_posts {
		post := Parse(&vk_posts[i], dictionary, organization)

		posts[post.Kind] = append(posts[post.Kind], *post)
	}

	return posts
}

func (posts Posts) Schedule() schedule.Schedule {
	s := make(schedule.Schedule, len(posts))
	for _, kind := range posts {
		for i := range kind {
			s = append(s, kind[i].Date)
		}
	}

	slices.SortFunc(s, func(a, b time.Time) int {
		if a.After(b) {
			return 1
		} else if a.Before(b) {
			return -1
		}

		return 0
	})

	return s
}
