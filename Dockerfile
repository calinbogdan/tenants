FROM golang:1.14.4-alpine AS builder

ENV TENANTSURL=undefined

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

# Build a small image
FROM scratch
COPY --from=builder /build/main ./

EXPOSE 5000

# Command to run
ENTRYPOINT ["/main"]