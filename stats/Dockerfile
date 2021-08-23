FROM golang:1.16.0-alpine3.13 AS builder
WORKDIR /app
COPY . .
RUN go build

FROM alpine:3.13
COPY --from=Builder /app/hugo-go-stats /bin/hugo-go-stats
CMD hugo-go-stats