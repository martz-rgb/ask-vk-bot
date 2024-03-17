package extract

import (
	"ask-bot/src/vk"
	"strings"
)

// int
func ID(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return message.ID
}

// string
func Attachments(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return vk.ToAttachments(message.Attachments)
}

// https://dev.vk.com/ru/reference/objects/photo-sizes
var photo_size_order = map[string]int{
	"":  -1,
	"o": -1,
	"p": -1,
	"q": -1,
	"r": -1,

	"s": 0,
	"m": 1,
	"x": 2,
	"y": 3,
	"z": 4,
	"w": 5,
}

// https://dev.vk.com/ru/reference/objects/photo-sizes#%D0%97%D0%BD%D0%B0%D1%87%D0%B5%D0%BD%D0%B8%D1%8F%20type%20%D0%B4%D0%BB%D1%8F%20%D0%B4%D0%BE%D0%BA%D1%83%D0%BC%D0%B5%D0%BD%D1%82%D0%BE%D0%B2%20(%D0%BF%D0%BE%D0%BB%D0%B5%20preview)
var doc_size_order = map[string]int{
	"":  -1,
	"s": 0,
	"m": 1,
	"x": 2,
	"y": 3,
	"z": 4,
	"o": 5,
}

// string
func Images(message *vk.Message) interface{} {
	var images []string

	for _, attachment := range message.Attachments {
		switch attachment.Type {
		case "photo":
			var image string
			current_size := ""

			for _, size := range attachment.Photo.Sizes {
				if photo_size_order[current_size] < photo_size_order[size.Type] {
					current_size = size.Type
					image = size.URL
				}
			}

			if len(image) != 0 {
				images = append(images, image)
			}
		case "doc":
			// https://dev.vk.com/ru/reference/objects/doc
			// Тип файла. Возможные значения:
			// ...
			// 4 — изображения;
			// ...

			if attachment.Doc.Type != 4 {
				continue
			}

			var image string
			current_size := ""

			for _, size := range attachment.Photo.Sizes {
				if photo_size_order[current_size] < doc_size_order[size.Type] {
					current_size = size.Type
					image = size.URL
				}
			}

			if len(image) != 0 {
				images = append(images, image)
			}
		}
	}

	return strings.Join(images, ",")
}
