FROM busybox
ADD git-sentinel /opt/git-sentinel
RUN mkdir /data
CMD ["/opt/git-sentinel"]