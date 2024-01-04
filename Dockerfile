FROM golang:1.21.3-alpine
WORKDIR /app/
COPY ./src go.mod go.sum schema.sql /app/

RUN apk add build-base icu-dev
RUN mkdir log && mkdir db
RUN CGO_ENABLED=1 go build -o bot -tags "icu" . 

CMD ["/app/bot"]