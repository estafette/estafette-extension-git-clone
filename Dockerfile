FROM alpine:3.6

MAINTAINER estafette.io

RUN apk add --no-cache \
    git

COPY estafette-extension-git-clone /

ENTRYPOINT ["/estafette-extension-git-clone"]