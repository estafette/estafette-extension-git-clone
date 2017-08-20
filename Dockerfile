FROM alpine:3.6

MAINTAINER estafette.io

RUN apk --no-cache add git

COPY estafette-extension-git-clone /

ENTRYPOINT ["/estafette-extension-git-clone"]