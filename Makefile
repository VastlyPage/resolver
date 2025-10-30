build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o server main.go

up:
	rsync -avP server root@hl.place:~/hlplace/upgraded
