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

TO-DO list:

- [ ] greetings (table + add from listener + create in postponed)
- [ ] greetings mode (manual or auto)
- [ ] notifications to admin (about new reservation if should be considerate and greetings if manual)
- [ ] roles groups + boards\albums
- [ ] points (spend)


thoughts and long-term:

- [ ] add timeslot test for exclude\merge
- [ ] use context in nodes
- [ ] print in log unsended 
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