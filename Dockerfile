FROM alpine:3.12

RUN addgroup -g 144 vanity && adduser -u 144 -S -G vanity vanity

RUN apk --update --no-cache add ca-certificates curl

COPY dist/go-vanity /bin/go-vanity
RUN chmod 755 /bin/go-vanity

RUN mkdir /vanity
