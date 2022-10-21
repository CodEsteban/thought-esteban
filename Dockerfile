FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR /usr/app

COPY go.mod .
COPY go.sum .
COPY thought.go .

# Build the binary.
RUN go get -d -v
RUN CGO_ENABLED=0 go build -o /usr/app/thought


FROM scratch

ENV PORT=3000
COPY --from=builder /usr/app/thought /usr/app/thought
ENTRYPOINT ["/usr/app/thought"]