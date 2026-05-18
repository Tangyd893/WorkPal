FROM golang:1.23-alpine AS builder

ARG SERVICE=gateway
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/workpal ./cmd/${SERVICE}

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /out/workpal /app/workpal
COPY configs /app/configs

EXPOSE 8080
CMD ["/app/workpal"]
