FROM golang:1.16.3-alpine3.13 AS builder
WORKDIR /go/src/github.com/simplesurance/registrator/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
		-a -installsuffix cgo \
		-ldflags "-X main.Version=$(cat VERSION)" \
		-o bin/registrator \
		.

FROM alpine:3.13
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/simplesurance/registrator/bin/registrator /bin/registrator

ENTRYPOINT ["/bin/registrator"]
