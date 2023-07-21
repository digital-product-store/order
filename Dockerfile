FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app

RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go generate ./...
RUN go build -v -o . ./...

FROM alpine:3.18

RUN mkdir -pv /opt/ads-order
WORKDIR /opt/ads-order

COPY --from=builder /usr/src/app/order .

USER nobody:nobody
CMD ["./order"]
