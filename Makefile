build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o server main.go

up:
	rsync -avP server hlplace@hl.place:~/hlplace/upgraded
