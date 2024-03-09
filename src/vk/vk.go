package vk

import (
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/awnumar/memguard"
)

// api.VK is based on http.Client and http.Client is claimed to be concurrency safe
type VK struct {
	id int

	api *api.VK
	r   *rand.Rand
}

func NewFromFile(name string, id int) (*VK, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token, err := memguard.NewBufferFromEntireReader(file)
	if err != nil {
		return nil, err
	}
	defer token.Destroy()

	if token.Size() == 0 {
		return nil, errors.New("group token is not provided")
	}

	api, err := New(id, token)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func New(id int, token *memguard.LockedBuffer) (*VK, error) {
	v := &VK{}
	source := rand.NewSource(time.Now().UnixNano())
	v.r = rand.New(source)
	v.id = id

	// should copy string because VK saves it inside and use,
	// but i destroy LockedBuffers with pointer on string
	v.api = api.NewVK(strings.Clone(token.String()))

	return v, nil
}

func (v *VK) ID() int {
	return v.id
}

func (v *VK) NewLongPoll() (*longpoll.LongPoll, error) {
	return longpoll.NewLongPoll(v.api, v.id)
}
