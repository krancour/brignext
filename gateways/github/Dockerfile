FROM krancour/go-tools:v0.1.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /src
COPY gateways/github/ gateways/github/
COPY pkg/ pkg/
COPY client/ client/
COPY *.go ./
COPY vendor/ vendor/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go build \
  -o bin/github-gateway \
  -ldflags "-w -X github.com/krancour/brignext/pkg/version.version=$VERSION -X github.com/krancour/brignext/pkg/version.commit=$COMMIT" \
  ./gateways/github/cmd

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brignext/bin/
ENTRYPOINT ["/brignext/bin/github-gateway"]
