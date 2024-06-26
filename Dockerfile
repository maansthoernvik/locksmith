# syntax=docker/dockerfile:1
FROM golang:1.21 AS build

# Set destination for COPY
WORKDIR /

COPY . .

ARG VERSION
ARG COMMIT
ENV VERSION=${VERSION}
ENV COMMIT=${COMMIT}

RUN CGO_ENABLED=0 GOOS=linux ./build-set-version locksmith ./exec/server

FROM alpine:latest

WORKDIR /

COPY --from=build /locksmith .

# Run
CMD ["./locksmith"]
