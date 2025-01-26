#!/bin/bash


sudo make clean -C ../../frostfs-aio
sudo make up -C ../../frostfs-aio 

sudo make refill WALLET=wallets/wallet1.json GAS=999999 -C ../../frostfs-aio/
sudo make refill WALLET=wallets/testwallet.json GAS=9999999 -C ../../frostfs-aio/


neo-go contract compile -i ./task_token/main.go -o ./task_token/main.nef -c ./task_token/neo-go.yml -m ./task_token/neo.manifest.json
neo-go contract deploy -i ./task_token/main.nef -m ./task_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 
neo-go contract generate-rpcwrapper -c ./task_token/neo-go.yml -m ./task_token/neo.manifest.json -o ../backend/wrappers/tasktoken/main.go --hash $task_hash


neo-go contract compile -i ./solution_token/main.go -o ./solution_token/main.nef -c ./solution_token/neo-go.yml -m ./solution_token/neo.manifest.json
neo-go contract deploy -i ./solution_token/main.nef -m ./solution_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 
neo-go contract generate-rpcwrapper -c ./solution_token/neo-go.yml -m ./solution_token/neo.manifest.json -o ../backend/wrappers/solutiontoken/main.go --hash $sol_hash

6a8f3f5454ca4f3830f2218dde3fd96723e46ecc
eabc5a8ad1d235c2bef03097c914a2c6c144213e


