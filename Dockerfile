FROM scratch
ADD ./ca-certificates.crt /etc/ssl/certs/
COPY ./gdipwebc /gdipwebc
CMD ["/gdipwebc"]
