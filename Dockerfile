FROM golang:1.20 as builder

WORKDIR /app

COPY . .

RUN if [ ! -f go.mod ]; then go mod init main; fi
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest

COPY --from=builder /app/main /app/

VOLUME "/app"

RUN ls

CMD ["/app/main", "/app/input.txt"]