################
# BUILD BINARY #
################
# golang:1.13.15-alpine3.12
FROM golang@sha256:0046254e021d5f1e26c844daac1ee406e382ac4a0f6c2e79593492ef1d90675f as builder


# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR $GOPATH/src/panorama
COPY . .

RUN echo $PWD && ls -lah

# Fetch dependencies.
# RUN go get -d -v
RUN go mod download
RUN go mod verify

#CMD go build -v
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/panorama .

#####################
# MAKE SMALL BINARY #
#####################
FROM alpine:3.12

RUN apk update

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy the executable.
COPY --from=builder /go/bin/panorama /go/bin/panorama
