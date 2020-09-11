FROM krancour/go-tools:v0.3.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /src
COPY observer/ observer/
COPY internal/ internal/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go build \
  -o bin/observer \
  -ldflags "-w -X github.com/brigadecore/brigade/v2/internal/version.version=$VERSION -X github.com/brigadecore/brigade/v2/internal/version.commit=$COMMIT" \
  ./observer

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brigade/bin/
ENTRYPOINT ["/brigade/bin/observer"]
