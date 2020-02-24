FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

RUN adduser -D -u 1001 -g '' appuser
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/api .

FROM scratch

WORKDIR /app

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/api /app/api
COPY .env /app/.env

ENTRYPOINT ["/app/api"]