# This is the first stage, for building things that will be required by the
# final stage (notably the binary)
FROM golang:1.19.2@sha256:0467d7d12d170ed8d998a2dae4a09aa13d0aa56e6d23c4ec2b1e4faacf86a813 AS builder

WORKDIR /go/src/app

# Copy in just the go.mod and go.sum files, and download the dependencies. By
# doing this before copying in the other dependencies, the Docker build cache
# can skip these steps so long as neither of these two files change.
COPY go.mod go.sum ./

# Assuming the source code is collocated to this Dockerfile
COPY . .

# Build the Go app with CGO_ENABLED=0 so we use the pure-Go implementations for
# things like DNS resolution (so we don't build a binary that depends on system
# libraries)
RUN CGO_ENABLED=0 go build -o /experia-v10-exporter

# Create a "nobody" non-root user for the next image by crafting an /etc/passwd
# file that the next image can copy in. This is necessary since the next image
# is based on scratch, which doesn't have adduser, cat, echo, or even sh.
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

# The second and final stage
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /experia-v10-exporter /experia-v10-exporter

# Copy the certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the /etc/passwd file we created in the builder stage. This creates a new
# non-root user as a security best practice.
COPY --from=builder /etc/passwd /etc/passwd

# Run as the new non-root by default
USER nobody

ENTRYPOINT [ "/experia-v10-exporter" ]
