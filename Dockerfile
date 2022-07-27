FROM golang:1.18.3-alpine3.16

WORKDIR /eggshellver
COPY . .
RUN go mod download
RUN apk add build-base
RUN apk add bash

RUN go build

CMD ["bash", "run.sh"]