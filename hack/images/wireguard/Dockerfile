ARG GOLANG_VERSION=1.19
ARG ALPINE_VERSION=3.16

FROM golang:${GOLANG_VERSION} as builder

ARG wg_go_tag=0.0.20220316
ARG wg_tools_tag=v1.0.20210914

RUN apt-get update -qq && apt-get install -yq git build-essential libmnl-dev iptables


RUN git clone https://git.zx2c4.com/wireguard-go && \
    cd wireguard-go && \
    git checkout $wg_go_tag && \
    make && \
    make install

ENV WITH_WGQUICK=yes
RUN git clone https://git.zx2c4.com/wireguard-tools && \
    cd wireguard-tools && \
    git checkout $wg_tools_tag && \
    cd src && \
    make && \
    make install



FROM debian:bookworm-slim


ENTRYPOINT ["/entrypoint.sh"]
WORKDIR /etc/wireguard

COPY entrypoint.sh /entrypoint.sh

COPY --from=builder /usr/bin/wireguard-go /usr/local/bin/wireguard-go
COPY --from=builder /usr/bin/wg-quick /usr/local/bin/wg-quick
COPY --from=builder /usr/bin/wg /usr/local/bin/wg

RUN apt-get update -qq && apt-get install -yq --no-install-suggests --no-install-recommends iproute2 iptables gettext-base fswatch

ENV SUB_NET=10.8.0.0/24
ENV WG_ERR_LOG_FILE=/dev/stderr
ENV WG_LOG_FILE=/dev/stdout
ENV WG_LOG_LEVEL=info
ENV WG_SUDO=1
ENV WG_THREADS=4
ENV WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go


CMD ["bash"]
