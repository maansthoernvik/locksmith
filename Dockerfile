# syntax=docker/dockerfile:1
FROM golang:1.21

# Set destination for COPY
RUN mkdir locksmith
RUN CGO_ENABLED=0 GOOS=linux go build -o /locksmith

# Download Go modules
#COPY go.mod go.sum ./
#RUN go mod download

# Run
CMD ["/locksmith"]
