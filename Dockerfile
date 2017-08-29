FROM alpine:3.6

LABEL maintainer="estafette.io" \
      description="The estafette-extension-git-clone component is an Estafette extension to clone a git repository for builds handled by Estafette CI"

RUN apk --no-cache add git

COPY estafette-extension-git-clone /

ENTRYPOINT ["/estafette-extension-git-clone"]