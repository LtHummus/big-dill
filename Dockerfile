FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM golang:1.25 AS build-go
COPY --from=xx / /

WORKDIR /go/src/app
COPY go.mod go.mod
COPY go.sum go.sum

RUN xx-go mod download

COPY . .
ENV CGO_ENABLED=0

ARG TARGETPLATFORM
RUN xx-go build -o /tmp/bigdill . && xx-verify /tmp/bigdill

FROM gcr.io/distroless/static-debian12:latest

ENV MODE=prod

COPY --from=build-go /tmp/bigdill /bin/bigdill

ENTRYPOINT ["bigdill"]
