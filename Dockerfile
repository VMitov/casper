FROM scratch
ADD ./casper /
RUN ["/casper"]
