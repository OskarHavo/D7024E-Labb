FROM golang:1.17.0-bullseye
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go mod download
RUN go build -o /hello-docker .
CMD ["/hello-docker"]
