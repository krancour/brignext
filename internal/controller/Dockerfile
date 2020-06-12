FROM krancour/go-tools:v0.1.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /src
COPY internal/controller/ internal/controller/
COPY internal/pkg/ internal/pkg/
COPY *.go ./
COPY vendor/ vendor/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go build \
  -o bin/controller \
  -ldflags "-w -X github.com/krancour/brignext/v2/internal/pkg/version.version=$VERSION -X github.com/krancour/brignext/v2/internal/pkg/version.commit=$COMMIT" \
  ./internal/controller/cmd

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brignext/bin/
ENTRYPOINT ["/brignext/bin/controller"]
