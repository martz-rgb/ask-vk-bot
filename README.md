about logging:
1) if there is an error from others code or there is new specific error, it should be wrapped
2) if there is an error from this project's parts (like *Ask, *VK and others), it should be taken as it is without wrapping

TO-DO list:

- [ ] db migration
- [ ] надо куда-то запихнуть админов/система прав
- [ ] manage secrets with docker swarm? 
- [ ] listener time probably
- [ ] or points (spend) or members\reservations

done:

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
