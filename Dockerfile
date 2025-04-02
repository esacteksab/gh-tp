FROM golang:1.24.1-bookworm AS builder

RUN apt update && apt install -y unzip wget git

RUN wget https://github.com/cli/cli/releases/download/v2.69.0/gh_2.69.0_linux_amd64.deb && dpkg -i gh_2.69.0_linux_amd64.deb && rm gh_2.69.0_linux_amd64.deb
RUN wget https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip && unzip terraform_1.5.7_linux_amd64.zip && rm terraform_1.5.7_linux_amd64.zip && mv terraform /usr/local/bin/terraform && chmod +x /usr/local/bin/terraform
RUN wget https://github.com/opentofu/opentofu/releases/download/v1.9.0/tofu_1.9.0_amd64.deb && dpkg -i tofu_1.9.0_amd64.deb && rm tofu_1.9.0_amd64.deb

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN scripts/build-dev.sh
RUN scripts/help-docker.sh

FROM builder AS test-stage

CMD [ "go", "test", "./...", "-cover"]
