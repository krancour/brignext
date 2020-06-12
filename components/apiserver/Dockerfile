FROM krancour/go-tools:v0.1.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /src
COPY components/apiserver/ components/apiserver/
COPY internal/ internal/
COPY meta/ meta/
COPY *.go ./
COPY vendor/ vendor/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go build \
  -o bin/apiserver \
  -ldflags "-w -X github.com/krancour/brignext/v2/internal/version.version=$VERSION -X github.com/krancour/brignext/v2/internal/version.commit=$COMMIT" \
  ./components/apiserver

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brignext/bin/
COPY components/apiserver/schemas/ /brignext/schemas/
ENTRYPOINT ["/brignext/bin/apiserver"]
