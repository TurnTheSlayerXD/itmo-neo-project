#!/bin/bash

#solution Contract: 7662b04f3dab029b7fcc99910a9599be5a1dfe7d

sudo make clear



neo-go contract compile -i ./solution_token/main.go -o ./solution_token/main.nef -c ./solution_token/neo-go.yml -m ./solution_token/neo.manifest.json
neo-go contract compile -i ./task_token/main.go -o ./task_token/main.nef -c ./task_token/neo-go.yml -m ./task_token/neo.manifest.json

neo-go contract deploy -i ./solution_token/main.nef --out ./solution_token/deploy.json -m ./solution_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 
neo-go contract deploy -i ./task_token/main.nef --out ./task_token/deploy.json -m ./task_token/neo.manifest.json -w ../wallets/wallet1.json -r http://localhost:30333 


