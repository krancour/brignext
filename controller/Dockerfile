FROM krancour/go-tools:v0.1.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/krancour/brignext
COPY controller/ controller/
COPY pkg/ pkg/
COPY client/ client/
COPY *.go ./
COPY vendor/ vendor/
RUN go build \
  -o bin/controller \
  -ldflags "-w -X github.com/krancour/brignext/pkg/version.version=$VERSION -X github.com/krancour/brignext/pkg/version.commit=$COMMIT" \
  ./controller/cmd

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/krancour/brignext/bin/ /brignext/bin/
ENTRYPOINT ["/brignext/bin/controller"]
