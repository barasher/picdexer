# build stage
FROM golang:1.15.7-alpine3.13 AS build-env
WORKDIR /root
RUN apk --no-cache add build-base git
ADD . /root
RUN go version
RUN go build -o picdexer picdexer.go

# final stage
FROM alpine
WORKDIR /root
RUN apk --no-cache add bash exiftool imagemagick
RUN mkdir -p /data/picdexer/in
RUN mkdir -p /data/picdexer/out
RUN mkdir -p /data/picdexer/dropzone
RUN mkdir -p /etc/picdexer
COPY docker/full.sh docker/setup.sh docker/dropzone.sh /root/
RUN chmod u+x full.sh setup.sh dropzone.sh
COPY docker/picdexer.json /etc/picdexer/picdexer.json
COPY --from=build-env /root/picdexer /root
CMD [ "/bin/bash" ]