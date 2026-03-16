FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    wget \
    && rm -rf /var/lib/apt/lists/*

# 下载Bitcoin Core（arm64版本，因为你是M芯片）
RUN wget https://bitcoincore.org/bin/bitcoin-core-28.0/bitcoin-28.0-aarch64-linux-gnu.tar.gz \
    && tar -xzf bitcoin-28.0-aarch64-linux-gnu.tar.gz \
    && cp bitcoin-28.0/bin/* /usr/local/bin/ \
    && rm -rf bitcoin-28.0*

RUN useradd -m bitcoin

USER bitcoin
WORKDIR /home/bitcoin

EXPOSE 18443 18444

CMD ["bitcoind", \
    "-regtest", \
    "-server", \
    "-rpcuser=user", \
    "-rpcpassword=pass", \
    "-rpcallowip=0.0.0.0/0", \
    "-rpcbind=0.0.0.0", \
    "-fallbackfee=0.0001", \
    "-txindex=1", \
    "-printtoconsole"]