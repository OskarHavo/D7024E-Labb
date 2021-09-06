FROM golang:1.17.0-bullseye

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN go build

ENTRYPOINT ["./Hello_comms"]