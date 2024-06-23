# syntax=docker/dockerfile:1
FROM golang:1.21 AS build

# Set destination for COPY
WORKDIR /

COPY . .

RUN CGO_ENABLED=0 GOOS=linux ./build-set-version locksmith ./exec/server

FROM alpine:latest

WORKDIR /

COPY --from=build /locksmith .

# Run
CMD ["./locksmith"]
