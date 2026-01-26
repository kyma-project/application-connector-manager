FROM --platform=$BUILDPLATFORM golang:1.25.6-alpine3.23 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/application-connector-manager/components/central-application-gateway

WORKDIR $DOCK_PKG_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GOFIPS140=v1.0.0 go build -a -installsuffix cgo -o applicationgateway ./cmd/applicationgateway


FROM --platform=$BUILDPLATFORM golang:1.25.6-alpine3.23 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/application-connector-manager/components/central-application-gateway/applicationgateway .

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

ENV GODEBUG=fips140=only,tlsmlkem=0
CMD ["/app/applicationgateway"]
