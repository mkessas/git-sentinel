FROM alpine
ADD git-sentinel /opt/git-sentinel
RUN mkdir /data
RUN apk add --no-cache git
CMD ["/opt/git-sentinel"]