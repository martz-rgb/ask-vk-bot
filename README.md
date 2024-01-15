to use docker-compose copy and run in **project directory**:
```
git clone https://github.com/coleifer/sqlite-web.git
```


TO-DO list:

- [ ] db migration
- [ ] надо куда-то запихнуть админов/система прав
- [ ] ask architecture somehow (listener + db, points\persons)
- [ ] manage secrets with docker swarm? 

done:

- [x] logging & error handling
- [x] docker but not all
- [x] word search (with icu) (yes!!!!)
- [x] db in mount
- [x] smaller image (multistage or something)
- [x] web client for sqlite (make it useful)
- [x] change keyboard without delete
- [x] cache expired deleting
