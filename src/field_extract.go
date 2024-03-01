package main

import "ask-bot/src/vk"

func ExtractID(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return message.ID
}

func ExtractAttachments(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return vk.ToAttachments(message.Attachments)
}
