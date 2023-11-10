# syntax=docker/dockerfile:1

FROM golang:1.21

# Set destination for COPY
WORKDIR /locksmith

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /locksmith

# Run
CMD ["/locksmith"]
