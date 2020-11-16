FROM golang:1.14-alpine AS build
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add git make openssl

WORKDIR /go/src/linkerd-disable-injection-mutation-webhook
COPY go* ./
RUN go get -v ./...
ADD . .
RUN make app

FROM alpine
RUN apk --no-cache add ca-certificates && mkdir -p /app
WORKDIR /app
COPY --from=build /go/src/linkerd-disable-injection-mutation-webhook/linkerd-disable-injection-mutation-webhook .
CMD ["/app/linkerd-disable-injection-mutation-webhook"]
