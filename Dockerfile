# syntax=docker/dockerfile:1
FROM golang:1.21 AS build

# Set destination for COPY
RUN pwd
RUN ls -al

WORKDIR /locksmith

RUN ls -al

COPY . ./

RUN ls -al
#RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /locksmith

FROM alpine:latest

WORKDIR /

COPY --from=build /locksmith .

# Run
CMD ["/locksmith"]
