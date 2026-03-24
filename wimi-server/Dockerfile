FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

COPY bin/WiMi-0.1.7.deb /tmp/WiMi-0.1.7.deb
RUN (dpkg -i /tmp/WiMi-0.1.7.deb || (apt-get update && apt-get install -f -y)) \
    && rm -f /tmp/WiMi-0.1.7.deb \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /usr/local/bin/WiMi

ENV LD_LIBRARY_PATH=/usr/local/bin/WiMi/libs
ENV LD_PRELOAD=/usr/local/bin/WiMi/libs/libjemalloc.so.2

ENTRYPOINT ["/entrypoint.sh"]
