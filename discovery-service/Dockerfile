FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go mod tidy
RUN go build -o discovery-service .

CMD ["./discovery-service", "239.255.255.250"]