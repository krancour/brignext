FROM krancour/go-tools:v0.4.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
# WORKDIR /src
# COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/scheduler/ scheduler/
COPY v2/internal/ internal/
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum
RUN go build \
  -o ../bin/scheduler \
  -ldflags "-w -X github.com/brigadecore/brigade/v2/internal/version.version=$VERSION -X github.com/brigadecore/brigade/v2/internal/version.commit=$COMMIT" \
  ./scheduler

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brigade/bin/
ENTRYPOINT ["/brigade/bin/scheduler"]
