FROM golang:1.20.5

RUN apt-get update && apt-get install -y curl

RUN apt-get update && \
    apt-get install -y ffmpeg

WORKDIR /go/src/worker

# Copy go.mod and go.sum files to the current directory
COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN mkdir -p /usr/local/bin
# Build the Worker and store it in /usr/local/bin
# This is the path where the Worker will be executed from
# Recursively copy all files from the current directory to the WORKDIR
RUN go build -o /usr/local/bin/worker

CMD ["worker"]