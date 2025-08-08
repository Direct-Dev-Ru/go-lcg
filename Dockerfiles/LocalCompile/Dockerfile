FROM --platform=${BUILDPLATFORM} golang:1.24.6-alpine3.22 AS build
ARG TARGETOS
ARG TARGETARCH
RUN apk add git && go install mvdan.cc/garble@latest
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN go mod download
COPY . .

# RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/go-lcg .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} garble  -literals -tiny build -ldflags="-w -s" -o /out/go-lcg .

FROM scratch AS bin-unix
COPY --from=build /out/go-lcg /lcg

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/go-lcg /lcg.exe

FROM bin-${TARGETOS} AS bin