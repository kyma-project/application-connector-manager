FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine3.22 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/central-application-connectivity-validator

WORKDIR $DOCK_PKG_DIR

COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:65534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o centralapplicationconnectivityvalidator ./cmd/centralapplicationconnectivityvalidator


FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/central-application-connectivity-validator/centralapplicationconnectivityvalidator .

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

CMD ["/app/centralapplicationconnectivityvalidator"]
