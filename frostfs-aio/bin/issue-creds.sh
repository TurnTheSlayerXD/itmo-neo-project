#!/bin/bash

initUser() {
  /usr/bin/frostfs-s3-authmate register-user \
    --wallet $WALLET_PATH \
    --rpc-endpoint http://localhost:30333 \
    --username $USERNAME \
    --contract-wallet /config/s3-gw-wallet.json >/dev/null 2>&1 && touch $WALLET_CACHE/$USERNAME
}

issueAWS() {
  /usr/bin/frostfs-s3-authmate issue-secret \
    --wallet $WALLET_PATH \
    --peer localhost:8080 \
    --gate-public-key $S3_GATE_PUBLIC_KEY \
    --container-placement-policy "REP 1"
}

S3_GATE_PUBLIC_KEY=$(neo-go wallet dump-keys -w /config/s3-gw-wallet.json | tail -1)
WALLET_PATH=/wallets/$2
if [[ -z "$2" ]]; then
  WALLET_PATH=/config/user-wallet.json
fi

WALLET_CACHE=/data/wallets
mkdir -p $WALLET_CACHE

USERNAME=$(echo $WALLET_PATH | md5sum | cut -d' ' -f1)
if [ ! -e $WALLET_CACHE/$USERNAME ]; then
  initUser
fi

if [ $1 == "s3" ]; then
  issueAWS
fi
