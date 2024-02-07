ARG GOLANG_VERSION=1.22
ARG ALPINE_VERSION=3.16

FROM golang:${GOLANG_VERSION} as builder

ARG wg_go_tag=0.0.20220316
ARG wg_tools_tag=v1.0.20210914

RUN apt-get update -qq && apt-get install -yq git build-essential libmnl-dev iptables


RUN git clone https://github.com/WireGuard/wireguard-go.git && \
    cd wireguard-go && \
    git checkout $wg_go_tag && \
    make && \
    make install

ENV WITH_WGQUICK=yes
RUN git clone https://github.com/WireGuard/wireguard-tools.git && \
    cd wireguard-tools && \
    git checkout $wg_tools_tag && \
    cd src && \
    make && \
    make install

FROM mindflavor/prometheus-wireguard-exporter:3.6.3 as exporter

FROM debian:bookworm-slim


WORKDIR /etc/wireguard

COPY --from=builder /usr/bin/wireguard-go /usr/local/bin/wireguard-go
COPY --from=builder /usr/bin/wg-quick /usr/local/bin/wg-quick
COPY --from=builder /usr/bin/wg /usr/local/bin/wg
COPY --from=exporter /usr/local/bin/prometheus_wireguard_exporter /usr/local/bin/prometheus_wireguard_exporter

RUN apt-get update -qq && apt-get install -yq --no-install-suggests --no-install-recommends iproute2 iptables gettext-base fswatch

ENV SUB_NET=10.8.0.0/24
ENV WG_ERR_LOG_FILE=/dev/stderr
ENV WG_LOG_FILE=/dev/stdout
ENV WG_LOG_LEVEL=info
ENV WG_SUDO=1
ENV WG_THREADS=4
ENV WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go

WORKDIR /
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

WORKDIR /etc/wireguard

CMD ["bash"]
