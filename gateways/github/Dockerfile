FROM quay.io/deis/lightweight-docker-go:v0.7.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/krancour/brignext
COPY gateways/github/ gateways/github/
COPY pkg/ pkg/
COPY client/ client/
COPY *.go ./
COPY vendor/ vendor/
RUN go build \
  -o bin/github-gateway \
  -ldflags "-w -X github.com/krancour/brignext/pkg/version.version=$VERSION -X github.com/krancour/brignext/pkg/version.commit=$COMMIT" \
  ./gateways/github/cmd

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/krancour/brignext/bin/ /brignext/bin/
ENTRYPOINT ["/brignext/bin/github-gateway"]
