FROM --platform=${BUILDPLATFORM} golang:1.24.6-alpine3.22 AS builder

ARG TARGETARCH

RUN apk add git 
#&& go install mvdan.cc/garble@latest

WORKDIR /app

COPY . .

RUN echo $BUILDPLATFORM > buildplatform
RUN echo $TARGETARCH > targetarch

RUN GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /app/go-lcg .
#RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} garble -literals -tiny build -ldflags="-w -s" -o /app/go-lcg .

FROM alpine:latest

WORKDIR /root

# COPY --from=builder /app/buildplatform .
# COPY --from=builder /app/targetarch .
COPY --from=builder /app/go-lcg /root/lcg

ENTRYPOINT ["/root/lcg"]