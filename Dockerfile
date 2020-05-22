# build stage
FROM golang:1.14.2-alpine3.11 AS build-env
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
COPY docker/binary_push.sh docker/binary_simulate.sh docker/full.sh docker/metadata_index.sh docker/metadata_simulate.sh docker/setup.sh docker/dropzone.sh /root/
RUN chmod u+x binary_push.sh binary_simulate.sh full.sh metadata_index.sh metadata_simulate.sh setup.sh dropzone.sh
COPY docker/picdexer.json /etc/picdexer/picdexer.json
COPY --from=build-env /root/picdexer /root
CMD [ "/bin/bash" ]