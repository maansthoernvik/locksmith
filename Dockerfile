# syntax=docker/dockerfile:1
FROM golang:1.21

# Set destination for COPY
RUN pwd
RUN ls -al

WORKDIR /locksmith

RUN ls -al

COPY server ./
COPY go.mod ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /locksmith

RUN ls -al

# Download Go modules
#COPY go.mod go.sum ./
#RUN go mod download

# Run
CMD ["/locksmith"]
