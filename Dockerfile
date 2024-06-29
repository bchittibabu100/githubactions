FROM alpine:latest

# Install dependencies
RUN apk add --no-cache \
    curl \
    tar \
    bash

# Install Docker CLI
RUN curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz | tar xz --strip-components=1 -C /usr/local/bin docker/docker

# Verify installation
RUN docker --version

# Default command to run Bash
CMD ["bash"]
