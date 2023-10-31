package main

import (
	"encoding/json"
	"io"
	"os"
)

// "github.com/SevereCloud/vksdk/v2/api"
// "github.com/SevereCloud/vksdk/v2/events"
// longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"

type Config struct {
	GroupToken string `json:"GROUP_TOKEN"`
	AdminToken string `json:"ADMIN_TOKEN"`
	DataBase   string `json:"DATABASE"`

	AppId        int64  `json:"APP_ID"`
	ProtectedKey string `json:"PROTECTED_KEY"`
	ServerKey    string `json:"SERVER_KEY"`
}

func main() {
	config_file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer config_file.Close()

	content, _ := io.ReadAll(config_file)

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		panic(err)
	}

	if len(config.GroupToken) == 0 {
		panic("no group token in config file")
	}
	if len(config.AdminToken) == 0 {
		panic("no admin token in config file")
	}
	if len(config.DataBase) == 0 {
		panic("no database url in config file")
	}

	api, err := NewVkApi(config.GroupToken, config.AdminToken)
	if err != nil {
		panic(err)
	}

	db, err := NewDb(config.DataBase)
	if err != nil {
		panic(err)
	}

	chat_bot := NewChatBot(dict, InitState, api, db)
	chat_bot.RunLongPoll()
}

////////////////////////////////////////////////

// postponed posts
// get_params := api.Params{
// 	"owner_id": (-1) * group[0].ID,
// 	"count":    10,
// 	"filter":   "postponed",
// }
// r, err := vk_wall.WallGet(get_params)
// if err != nil {
// 	panic(err)
// }
// fmt.Println("postponed result", r)

// запостим нафиг
// post_params := api.Params{
// 	"owner_id": (-1) * group[0].ID,
// 	"post_id":  r.Items[0].ID,
// }
// answer, err := vk_wall.WallPost(post_params)
// if err != nil {
// 	panic(err)
// }
// fmt.Println(answer)

// delete keyboard from message with empty text
// edit_params := api.Params{
// 	"peer_id":                 obj.PeerID,
// 	"conversation_message_id": obj.ConversationMessageID,
// 	"message":                 "Техническое сообщение",
// 	"keyboard":                EmptyKeyboard,
// }
// response, err = vk_group.MessagesEdit(edit_params)
// if err != nil {
// 	fmt.Println("err", err)
// 	return
// }
// fmt.Println("response", response)

//   source := rand.NewSource(time.Now().UnixNano())
// 	random := rand.New(source)

// 	vk := api.NewVK(config.VkToken)

// 	// Получаем информацию о группе
// 	group, err := vk.GroupsGetByID(api.Params{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("OK", group[0].ID)

// lp, err := longpoll.NewLongPoll(vk, group[0].ID)
// if err != nil {
// 	panic(err)
// }

// lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
// 	fmt.Printf("%+v\n", obj.Message)
// 	params := api.Params{
// 		"user_id":   obj.Message.FromID,
// 		"random_id": random.Int(),
// 		"message":   "Я получил ваше сообщение.",
// 		"keyboard":  Keyboard,
// 		"group_id":  group[0].ID,
// 		"peer_id":   []int{obj.Message.FromID},
// 	}
// 	response, err := vk.MessagesSend(params)
// 	if err != nil {
// 		fmt.Println("err", err)
// 		return
// 	}
// 	fmt.Println("response", response)

// })
// lp.MessageEvent(func(ctx context.Context, obj events.MessageEventObject) {
// 	fmt.Printf("message event %+v\n", obj)
// 	// answer on callback to clear loading
// 	params := api.Params{
// 		"event_id":  obj.EventID,
// 		"user_id":   obj.UserID,
// 		"random_id": random.Int(),
// 		"message":   "Я тоже надеюсь, что вы им станете",
// 		"peer_id":   []int{obj.PeerID},
// 	}
// 	response, err := vk.MessagesSendMessageEventAnswer(params)
// 	if err != nil {
// 		fmt.Println("err", err)
// 		return
// 	}
// 	fmt.Println("response", response)

// 	// send actual message
// 	params = api.Params{
// 		"user_id":   obj.UserID,
// 		"random_id": random.Int(),
// 		"message":   "Я тоже хотел бы.",
// 		// "keyboard":  Keyboard,
// 		"group_id": group[0].ID,
// 	}
// 	response, err = vk.MessagesSend(params)
// 	if err != nil {
// 		fmt.Println("err", err)
// 		return
// 	}
// 	fmt.Println("response", response)

// })

// lp.Run()

// to-do попробовать Exectute?
// интерестинг, но того не стоит
// var execute_script string = `
// 	var a = API.messages.send({"user_id": %d, "random_id": %d, "message": "%s", "keyboard": "%s"});
// 	return API.messages.delete({"peer_id": %d, "message_ids": a, "delete_for_all": 1});
// `

// var del_response api.MessagesDeleteResponse

// script := fmt.Sprintf(execute_script,
// 	obj.UserID,
// 	m.r.Int(),
// 	"Меняю клавиатуру...",
// 	"",
// 	obj.PeerID)

// fmt.Println((script))
// err = m.group.Execute(script, &del_response)

// if err != nil {
// 	fmt.Println(err)
// }

// send actual message
// params = api.Params{
// 	"user_id":   obj.UserID,
// 	"random_id": m.r.Int(),
// 	"message":   "Меняю клавиатуру...",
// 	"keyboard":  EmptyKeyboard,
// }
// response, err = m.group.MessagesSend(params)
// if err != nil {
// 	fmt.Println("err", err)
// 	return
// }

// fmt.Println("response", response)

// var Keyboard = `{
// 	"buttons": [
// 	  [
// 		{
// 		  "action": {
// 			"type": "open_link",
// 			"link": "https://dev.vk.com/",
// 			"label": "Оформить"
// 		  }
// 		},
// 		{
// 		  "action": {
// 			"type": "callback",
// 			"label": "Хочу стать отвечающим",
// 			"payload": "{}"
// 		  }
// 		}
// 	  ]
// 	]
//   }`

// var EmptyKeyboard = `{
// 	"buttons": []
// }`
