FROM golang:1.23

WORKDIR /src/pinger

COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "run", "-race", "cmd/main.go"]