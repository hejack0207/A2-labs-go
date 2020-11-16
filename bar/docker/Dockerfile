FROM centos:7

COPY ./bar /bar/
RUN chmod u+x /bar/bar

EXPOSE 8080
# CMD tail -f /dev/null
CMD ["/bar/bar"]

