FROM busybox
ADD git-sentinel /opt/git-sentinel
ADD ca-certificates.crt /etc/ssl/certs/
RUN mkdir /data
CMD ["/opt/git-sentinel"]