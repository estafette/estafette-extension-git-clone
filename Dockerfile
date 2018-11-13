FROM alpine:3.8

LABEL maintainer="estafette.io" \
      description="The estafette-extension-git-clone component is an Estafette extension to clone a git repository for builds handled by Estafette CI"

RUN apk add --update --no-cache \
    git \
    && rm -rf /var/cache/apk/*

COPY estafette-extension-git-clone /

ENTRYPOINT ["/estafette-extension-git-clone"]