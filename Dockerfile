FROM golang:1.21.3-alpine

WORKDIR /build
RUN apk add build-base icu-dev

COPY ./src go.mod go.sum ./
RUN go mod download
RUN go build -o app -tags "icu" .

FROM alpine:3

WORKDIR /bin

COPY schema.sql .
COPY --from=0 /build/app main
COPY --from=0 /usr/lib/libicuuc.so.73 /usr/lib/libicui18n.so.73 /usr/lib/libicudata.so.73 /usr/lib/libstdc++.so.6  /usr/lib/libgcc_s.so.1 /usr/lib/

CMD ["/bin/main"]