FROM alpine:3.7

LABEL maintainer="estafette.io" \
      description="The estafette-extension-git-clone component is an Estafette extension to clone a git repository for builds handled by Estafette CI"

RUN apk --update --no-cache --virtual add git && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

COPY estafette-extension-git-clone /

ENTRYPOINT ["/estafette-extension-git-clone"]