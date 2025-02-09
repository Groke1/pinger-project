FROM golang:1.23

WORKDIR /src/backend

COPY . .

RUN go mod tidy

EXPOSE 8080

ENTRYPOINT ["go", "run", "-race", "cmd/main.go"]