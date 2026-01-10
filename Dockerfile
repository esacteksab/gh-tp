FROM esacteksab/go:1.25.5-2026-01-09@sha256:d587e1b26479b4591e6c38d8f67afff5dd41c86c57698b5134fab0590d5f351b AS builder

# Set GOMODCACHE explicitly (still good practice)
ENV GOMODCACHE=/go/pkg/mod

# Keep this layer cached if possible
RUN apt update && apt install -y unzip wget git \
  && wget https://github.com/cli/cli/releases/download/v2.82.1/gh_2.82.1_linux_amd64.deb \
  && dpkg -i gh_2.82.1_linux_amd64.deb && rm gh_2.82.1_linux_amd64.deb \
  && wget https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip \
  && unzip terraform_1.5.7_linux_amd64.zip && rm terraform_1.5.7_linux_amd64.zip \
  && mv terraform /usr/bin/terraform && chmod +x /usr/bin/terraform \
  && wget https://github.com/opentofu/opentofu/releases/download/v1.9.0/tofu_1.9.0_amd64.deb \
  && dpkg -i tofu_1.9.0_amd64.deb && rm tofu_1.9.0_amd64.deb

WORKDIR /app

# Copy only module files first to maximize caching
COPY go.mod go.sum ./

# Download modules. This layer will be cached if go.mod/go.sum haven't changed.
# The downloaded files will now be part of this layer's filesystem.
RUN go mod download

# Copy the rest of the application code
COPY . .

# Keep cache mounts here for build performance (Go build cache + reusing modules during build)
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build scripts/build-dev.sh
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build scripts/help-docker.sh

# --- Test Stage ---
FROM builder AS test-stage

# No need to set GOMODCACHE again, inherited from builder
# No need to set WORKDIR again, inherited from builder

RUN mkdir -p /app/coverdata
ENV GOCOVERDIR=/app/coverdata

# Go test should now find modules in /go/pkg/mod inherited from the builder stage
CMD ["/bin/sh", "-c", "go test -covermode=atomic -coverprofile=/app/coverdata/coverage.out ./... && echo 'Coverage data collected'"]
