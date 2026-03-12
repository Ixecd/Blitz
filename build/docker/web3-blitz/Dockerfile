# Copyright 2025 qc <2192629378@qq.com>. All Rights Reserved.
# Use of this source code is governed by a MIT style
# License that can be found in the LICENSE file.

FROM BASE_IMAGE
LABEL maintainer="<2192629378@qq.com>"

WORKDIR /opt/iam

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
      echo "Asia/Shanghai" > /etc/timezone && \
      mkdir -p /var/log/iam

COPY web3-blitz /opt/web3-blitz/bin/

ENTRYPOINT ["/opt/web3-blitz/bin/web3-blitz"]
CMD ["-c", "/etc/web3-blitz/web3-blitz.yaml"]
