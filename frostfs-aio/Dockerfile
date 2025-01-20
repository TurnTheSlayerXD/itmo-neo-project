ARG FROSTFS_HUB_IMAGE=truecloudlab/frostfs
ARG NEOGO_HUB_IMAGE=nspccdev/neo-go

ARG FROSTFS_CORE_TAG=dev
ARG FROSTFS_HTTP_GATE_TAG=dev
ARG FROSTFS_S3_GATE_TAG=dev
ARG NEOGO_TAG=dev

FROM ${NEOGO_HUB_IMAGE}:${NEOGO_TAG} AS neo-go
FROM ${FROSTFS_HUB_IMAGE}-adm:${FROSTFS_CORE_TAG} AS frostfs-adm
FROM ${FROSTFS_HUB_IMAGE}-cli:${FROSTFS_CORE_TAG} AS frostfs-cli
FROM ${FROSTFS_HUB_IMAGE}-ir:${FROSTFS_CORE_TAG} AS frostfs-ir
FROM ${FROSTFS_HUB_IMAGE}-storage:${FROSTFS_CORE_TAG} AS frostfs-storage
FROM ${FROSTFS_HUB_IMAGE}-s3-gw:${FROSTFS_S3_GATE_TAG} AS frostfs-s3-gw
FROM ${FROSTFS_HUB_IMAGE}-http-gw:${FROSTFS_HTTP_GATE_TAG} AS frostfs-http-gw

# Executable image
FROM alpine AS frostfs-aio
RUN apk add --no-cache \
  bash \
  ca-certificates \
  jq \
  expect \
  iputils \
  curl

WORKDIR /

COPY --from=neo-go /usr/bin/neo-go /usr/bin/neo-go
COPY --from=frostfs-adm /bin/frostfs-adm /usr/bin/frostfs-adm
COPY --from=frostfs-cli /bin/frostfs-cli /usr/bin/frostfs-cli
COPY --from=frostfs-ir /bin/frostfs-ir /usr/bin/frostfs-ir
COPY --from=frostfs-storage /bin/frostfs-node /usr/bin/frostfs-node
COPY --from=frostfs-s3-gw /bin/frostfs-s3-gw /usr/bin/frostfs-s3-gw
COPY --from=frostfs-s3-gw /bin/frostfs-s3-authmate /usr/bin/frostfs-s3-authmate
COPY --from=frostfs-http-gw /bin/frostfs-http-gw /usr/bin/frostfs-http-gw

COPY ./adm/frostfs-adm.yml /config/frostfs-adm.yml
COPY ./ir/cli-cfg.yaml /config/cli-cfg-ir.yaml
COPY ./ir/config.yaml /config/config-ir.yaml
COPY ./morph/protocol.privnet.yml /config/protocol.privnet.yml
COPY ./morph/node-wallet.json /config/alphabet/az.json
COPY ./morph/node-wallet.json /config/alphabet/contract.json
COPY ./morph/node-wallet.json /config/node-wallet.json
COPY ./morph/node-config.yaml /config/node-config.yaml
COPY ./http-gw/http-gw-config.yaml /config/http-gw-config.yaml
COPY ./http-gw/http-gw-wallet.json /config/http-gw-wallet.json
COPY ./s3-gw/rules.json /config/bearer-rules.json
COPY ./s3-gw/regions.json /config/regions.json
COPY ./s3-gw/s3-gw-config.yaml /config/s3-gw-config.yaml
COPY ./s3-gw/s3-gw-wallet.json /config/s3-gw-wallet.json
COPY ./s3-gw/s3-gw-tls.crt /config/s3-gw-tls.crt
COPY ./s3-gw/s3-gw-tls.key /config/s3-gw-tls.key
COPY ./s3-gw/user-wallet.json /config/user-wallet.json
COPY ./sn/cli-cfg.yaml /config/cli-cfg-sn.yaml
COPY ./sn/wallet.json /config/wallet-sn.json
COPY ./sn/config.yaml /config/config-sn.yaml
COPY ./vendor/locode_db /config/locode.db
COPY ./vendor/contracts/ /config/contracts

COPY ./bin/init-aio.sh /usr/bin/init-aio.sh
COPY ./bin/issue-creds.sh /usr/bin/issue-creds.sh

ENV AUTHMATE_WALLET_PASSPHRASE=""
ENV AUTHMATE_WALLET_CONTRACT_PASSPHRASE=""

ENTRYPOINT ["/usr/bin/init-aio.sh"]
