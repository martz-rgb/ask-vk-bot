about logging:
1) if there is an error from others code or there is new specific error, it should be wrapped
2) if there is an error from this project's parts (like *Ask, *VK and others), it should be taken as it is without wrapping

about timezone:
даты в базе данных лежат в UTC и не зависят от часового пояса аска. Слой аска же выдает даты в нужном часовом поясе.

about reservation:
не больше одной брони на человека?

TO-DO list:

- [ ] polls
- [ ] listener time probably
- [ ] points (spend)
- [ ] roles groups + boards\albums
- [ ] configuration with or without confirmation reservations
- [ ] use context in nodes
- [ ] manage secrets with docker swarm? 
- [ ] longpoll quit silently if no internet connection
- [ ] *maybe some collaborators support


done:

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