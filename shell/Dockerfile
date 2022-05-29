FROM alpine:3.15
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing proot wget tar

RUN adduser -u 1000 -s /bin/sh -D smroot

USER smroot

WORKDIR /home/smroot

RUN  wget -O /tmp/alpine.tar.gz https://dl-cdn.alpinelinux.org/alpine/v3.15/releases/x86_64/alpine-minirootfs-3.15.4-x86_64.tar.gz

RUN mkdir /home/smroot/rootfs && tar -xvf /tmp/alpine.tar.gz -C /home/smroot/rootfs && \
    rm -rf /tmp/alpine.tar.gz


CMD ["/bin/sh", "-c", "proot -r /home/smroot/rootfs"]
