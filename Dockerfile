FROM golang:1.16-buster as builder

WORKDIR /go/src

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -o natalya.exe

FROM debian:buster-slim

WORKDIR /go/bin

RUN mkdir /assets
COPY assets/ /assets/

COPY --from=builder /go/src/natalya.exe /go/bin/app.exe
RUN chmod +x /go/bin/app.exe
COPY assets/serifs.yml /assets/serifs.yml

CMD ["/bin/bash", "-c", "/go/bin/app.exe"]

EXPOSE 8000
