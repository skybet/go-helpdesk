FROM golang:1.9-alpine AS builder

ENV BUILDROOT /go/src/github.com/skybet/go-helpdesk
ADD . $BUILDROOT
WORKDIR $BUILDROOT

RUN go test -v ./...; \
    go build .


FROM alpine
ENV BUILDROOT /go/src/github.com/skybet/go-helpdesk
COPY --from=builder $BUILDROOT/go-helpdesk /bin

EXPOSE 4390

ENTRYPOINT ["/bin/go-helpdesk"]
