Сервис-помощник в управлении творческими сообществами VK в формате «спроси персонажа».

Бот работает в режиме сервера и может (в светлом будущем):
- самостоятельно принимать брони и приветствия, публиковать опросы и принимать отвечающих на роль,
- обрабатывать ответы на публикацию, 
- подсчитывать дедлайны, предупреждать об их скором наступлении и оформлять уходы, 
- загружать ответы в альбом и создавать новые обсуждения, 
- поддерживать балловую систему в сообществе.

В этом проекте я стараюсь учесть все устоявшиеся привычки в ask-сообществах и создать наиболее гибкую и удобную систему, которая будет помогать администрации ask-сообществ в управлении с уменьшением рутины.

================

рабочее

about logging:
1) if there is an error from others code or there is new specific error, it should be wrapped
2) if there is an error from this project's parts (like *Ask, *VK and others), it should be taken as it is without wrapping

about timezone:
даты в базе данных лежат в UTC формате, но означают время в таймзоне аска. Аск сам пересчитывает даты, чтобы они были правильные в реальном мире

about reservation:
не больше одной брони на человека?

about back:
стандартное поведение при кнопке back -- заново запустить entry

about poll:
только одна роль может быть на одном опросе

МОЖЕТ БЫТЬ, имеет смысл вариант "урезанного" бота, который можно включать только раз в определенное время
скорее всего, этот вариант без основного функционала чатбота, способность добавлять самостоятельно брони и тд
думай.


TO-DO list:

- [ ]  add cleaning of poll answer cache
- [ ] greetings (table + add from listener + create in postponed)
- [ ] members & deadlines
- [ ] greetings mode (manual or auto)
- [ ] notifications to admin (about new reservation if should be considerate and greetings if manual)
- [ ] roles groups
- [ ] album\boards order?
- [ ] points (spend)
- [ ] faq (либо один текст, либо параметризация списка, либо просто выпилить нафиг)


thoughts and long-term:

- [ ] maybe remove error from ask.Schedule & create init method to check 
- [ ] postponed method more optimized? 
- [ ] stateless chatbot? maybe?
- [ ] some error resistent like not to stop all things but only affected by error (watcher)
- [ ] add timeslot test for exclude\merge
- [ ] use context in nodes
- [ ] print in log unsended 
- [ ] add prefix to bot messages
- [ ] flashmobs (create, submit, retrieve all, close)
- [ ] manage secrets with docker swarm? 
- [ ] longpoll quit silently if no internet connection
- [ ] какая-то фигня при смене клавиатуры на пк, если есть прикрепленное сообщение
- [ ] some env package to easily parse?
- [ ] *maybe some collaborators support (or maybe not)


done:

- [x] db migrate drop views/triggers associated with removing table
- [x] nodes have own payload
- [x] timestamp timezone by default is gmt+0, should be gmt+3 probably? -> timezone ask config
- [x] logging & error handling
- [x] docker but not all
- [x] word search (with icu) (yes!!!!)
- [x] db in mount
- [x] smaller image (multistage or something)
- [x] web client for sqlite (make it useful)
- [x] change keyboard without delete
- [x] cache expired deleting
- [x] um very thing to do IS TO REFACTOR NODES... i appreciate myself but what the f*ck (сойдет)
- [x] логирование без синглтона... -> no log in db\vk\nodes, up to chat; not in bot because i want it clear (i think?)
- [x] db migration (let's fucking go)
- [x] resetvations
- [x] система прав (пока достаточно)
- [x] deadline + member system 
- [x] notify users without breaking flow (notify only when stack is empty?)
- [x] postponed posts are handled
- [x] use context properly
- [x] db polls
- [x] create polls
- [x] listen to new polls
- [x] configuration with or without confirmation reservations
- [x] polls checking
- [x] reservation deadlines
- [x] boards\albums
- [x] polls sync
- [x] watcher with update on notifications
- [x] rewrite forms (make it more simple but save as mush power as can)
- [x] remember polls answers
