
FROM golang:1.13 as builder

RUN mkdir /app

ADD . /app/

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]

