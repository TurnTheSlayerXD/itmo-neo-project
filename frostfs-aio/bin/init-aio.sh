#!/bin/bash

stage() {
  echo "================================================================================"
  echo "$@"
  echo "================================================================================"
}

die() {
  echo "$@" 1>&2
  exit 1
}

runBlockchain() {
  stage "Starting the blockchain"

  /usr/bin/neo-go node --config-path /config --privnet |& tee -a ${LOG_DIR}/neo-go &

  while [[ "$(curl -s -o /dev/null -w %{http_code} localhost:30333)" != "422" ]];
  do
    sleep 2;
  done
}

configure() {
  stage "Configuring the blockchain"

  /usr/bin/frostfs-adm morph init --config /config/frostfs-adm.yml --contracts /config/contracts || die "Failed to initialize Alphabet wallets"

  /usr/bin/frostfs-adm --config /config/frostfs-adm.yml morph ape add-rule-chain --target-type namespace --target-name "" --rule 'allow Container.* *' --chain-id "allow_container_ops" || die "Failed to set defaul policy"

  /usr/bin/frostfs-adm morph refill-gas --config /config/frostfs-adm.yml --storage-wallet /config/wallet-sn.json --gas 10.0 || die "Failed to transfer GAS to alphabet wallets"

  /usr/bin/frostfs-adm morph proxy-add-account --config /config/frostfs-adm.yml --account NejLbQpojKJWec4NQRMBhzsrmCyhXfGJJe || die "Failed to set storage wallet as proxy wallet"

  /usr/bin/frostfs-adm morph proxy-add-account --config /config/frostfs-adm.yml --account NN1RQ3qwnvDMVNsQw4WPkKi7BrjxdVTDZp || die "Failed to set s3 gateway wallet as proxy wallet"
}

runServices() {
  stage "Running services"

  /usr/bin/frostfs-ir --config /config/config-ir.yaml |& tee -a ${LOG_DIR}/frostfs-ir &

  while [[ -z "$(/usr/bin/frostfs-cli control ir healthcheck --endpoint localhost:16512 -c /config/cli-cfg-ir.yaml | grep 'Health status: READY')" ]];
  do
    sleep 2;
  done

  set -m
  /usr/bin/frostfs-node --config /config/config-sn.yaml |& tee -a ${LOG_DIR}/frostfs-node &

  while [[ -z "$(/usr/bin/frostfs-cli control healthcheck --endpoint localhost:16513 -c /config/cli-cfg-sn.yaml | grep 'Health status: READY')" ]];
  do
    sleep 2
  done

  while [[ -z "$(/usr/bin/frostfs-cli control healthcheck --endpoint localhost:16513 -c /config/cli-cfg-sn.yaml | grep 'Network status: ONLINE')" ]];
  do
    /usr/bin/frostfs-adm morph force-new-epoch --config /config/frostfs-adm.yml || die "Failed to update epoch"
    sleep 2
  done

  while [[ -z "$(/usr/bin/frostfs-cli tree healthcheck -r 127.0.0.1:8080 -g -v | grep 'Successful healthcheck invocation')" ]];
  do
    sleep 2
  done

  /usr/bin/frostfs-s3-gw --config /config/s3-gw-config.yaml |& tee -a ${LOG_DIR}/frostfs-s3-gw &
  /usr/bin/frostfs-http-gw --config /config/http-gw-config.yaml |& tee -a ${LOG_DIR}/frostfs-http-gw &
}


if [ ! -e "/data/chain/morph.bolt" ];
then
  runBlockchain
  configure
else
  runBlockchain
fi
runServices
stage "aio container started"
fg
