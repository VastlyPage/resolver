build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o server main.go

up:
	rsync -avP server root@hl.baby:~/hlbaby/upgraded

renew:
	certbot certonly --manual --preferred-challenge dns -v -d *.hl.place,hl.place,*.hl.baby,hl.baby

certup:
	cp /etc/letsencrypt/live/hl.baby/fullchain.pem .
	cp /etc/letsencrypt/live/hl.baby/privkey.pem .
	rsync -avP fullchain.pem root@hl.baby:~/hlbaby/
	rsync -avP privkey.pem root@hl.baby:~/hlbaby/
