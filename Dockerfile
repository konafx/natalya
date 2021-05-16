FROM golang:1.16-buster as builder

WORKDIR /go/src

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -o natalya

FROM debian:buster-slim

WORKDIR /go/bin

ARG DISCORD_BOT_TOKEN
ARG DISCORD_GUILD_ID

COPY --from=builder /go/src/natalya /go/bin/app
RUN chmod +x /go/bin/app

CMD ["/bin/bash", "-c", "./app --guild ${DISCORD_GUILD_ID} --token ${DISCORD_BOT_TOKEN}"]

EXPOSE 8080
