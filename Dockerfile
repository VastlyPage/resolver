FROM debian:12-slim

RUN apt update
RUN apt upgrade -y
RUN apt install git curl wget -y

RUN useradd -ms /bin/bash user
USER user
WORKDIR /home/user

RUN wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz -O go.tar.gz
RUN tar xzf go.tar.gz

# Compile sources
COPY go.mod /home/user/
COPY go.sum /home/user/
RUN /home/user/go/bin/go mod download

COPY backend/ /home/user/backend/
COPY hlnames/ /home/user/hlnames/
COPY util/ /home/user/util/
COPY main.go /home/user/
RUN CGO_ENABLED=0 /home/user/go/bin/go build -ldflags="-s -w" -o /home/user/server main.go

ENTRYPOINT [ "/home/user/server" ]