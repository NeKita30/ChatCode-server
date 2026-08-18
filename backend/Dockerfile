FROM golang:1.25-alpine

RUN apk add --no-cache wget

WORKDIR /app
COPY ./src .
RUN go mod download

RUN go build .

EXPOSE 8080

ENTRYPOINT ["./chat_server"]
