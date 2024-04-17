FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct


WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .

// RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./main.go



FROM alpine:latest



WORKDIR /app
COPY --from=builder /app/main /app/main
COPY templates /app/templates

EXPOSE 8080

CMD ["./main"]
