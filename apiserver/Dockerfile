FROM quay.io/deis/lightweight-docker-go:v0.7.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/krancour/brignext
COPY apiserver/ apiserver/
COPY pkg/ pkg/
COPY *.go ./
COPY vendor/ vendor/
RUN go build \
  -o bin/apiserver \
  -ldflags "-w -X github.com/krancour/brignext/pkg/version.version=$VERSION -X github.com/krancour/brignext/pkg/version.commit=$COMMIT" \
  ./apiserver/cmd

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/krancour/brignext/bin/ /brignext/bin/
ENTRYPOINT ["/brignext/bin/apiserver"]
