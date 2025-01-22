#!/bin/bash

#solution Contract: 7662b04f3dab029b7fcc99910a9599be5a1dfe7d

sudo make clean -C ../../frostfs-aio
sudo make up -C ../../frostfs-aio 

neo-go contract compile -i ./solution_token/main.go -o ./solution_token/main.nef -c ./solution_token/neo-go.yml -m ./solution_token/neo.manifest.json
neo-go contract compile -i ./task_token/main.go -o ./task_token/main.nef -c ./task_token/neo-go.yml -m ./task_token/neo.manifest.json

task_hash=$(neo-go contract deploy -i ./task_token/main.nef --out ./task_token/deploy.json -m ./task_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 | grep 'Contract' | cut -d ' ' -f 2)
sol_hash=$(neo-go contract deploy -i ./solution_token/main.nef --out ./solution_token/deploy.json -m ./solution_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 $task_hash | grep 'Contract' | cut -d ' ' -f 2)

 neo-go contract generate-rpcwrapper -c ./solution_token/neo-go.yml -m ./solution_token/neo.manifest.json -o ../backend/wrappers/solutiontoken/main.go --hash $sol_hash
 neo-go contract generate-rpcwrapper -c ./task_token/neo-go.yml -m ./task_token/neo.manifest.json -o ../backend/wrappers/tasktoken/main.go --hash $task_hash


echo "sol hash: ${sol_hash}"
echo "task hash: ${task_hash}"


