# syntax=docker/dockerfile:1
FROM golang:1.21 AS build

# Set destination for COPY
WORKDIR /locksmith

COPY . ./

RUN ./build_with_parameters

FROM alpine:latest

WORKDIR /

COPY --from=build /locksmith .

# Run
CMD ["/locksmith"]
