#
# Container to build deploy static frontend over ssh.
# For use in Pipeline's.
#
# Software installed:
# - frontend-deploy
#
FROM alpine:latest
MAINTAINER Jason Lentink <jason@mediamonks.com>

ADD frontend-deploy /usr/bin/frontend-deploy


#
# Run install script
#
RUN  echo "Installing..." \
     && apk add --no-cache git \
     && git config --global --add safe.directory '*'
WORKDIR /app

ENTRYPOINT frontend-deploy


