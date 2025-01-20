
neo-go contract compile -i contracts/storage/counter.go -o contracts/storage/counter.nef -m contracts/storage/counter.manifest.json -c contracts/storage/counter.yml  


neo-go contract deploy -i contracts/hello-contract/hello.nef -m  contracts/hello-contract/hello.manifest.json -r http://localhost:30333 -w wallets/wallet1.json

d30ee5f6ada2f976d27a574a8a0bfca420edb20e5efa390c5d0b80cc426b4648


Sending transaction...


54efcf759c7b73dce6a5e340126e39fc1230f890c574f5e43920cda602cc84ed
Contract: 2a77d599f1172e2320f2f8e00e27c26a6209c8be


neo-go contract invokefunction -r http://localhost:30333 -w wallets/wallet1.json 2a77d599f1172e2320f2f8e00e27c26a6209c8be runtimeNotify [ string ]

8e92d337e17deea3d3896970935d34f1ea642fcc3b1cd4bd48ea71523100093c

curl -s --data '{"id":1,"jsonrpc":"2.0","method":"getapplicationlog","params":["8e92d337e17deea3d3896970935d34f1ea642fcc3b1cd4bd48ea71523100093c"]}' http://localhost:30333 



6cb3b49abfd004ddb1ac8c96e87933387484c57c9d0c63b8e3b6e1d51063c6a5
Contract: d411059af8ac1ba8268751867d873a1928ee8369
curl -s --data '{"id":1,"jsonrpc":"2.0","method":"getapplicationlog","params":["fe924b7cfe89ddd271abaf7210a80a7e11178758"]}' http://localhost:30333 

neo-go contract invokefunction -r http://localhost:30333 -w wallets/wallet1.json d411059af8ac1ba8268751867d873a1928ee8369 main



neo-go contract compile -i contracts/proxygetter/proxygetter.go -o contracts/proxygetter/proxygetter.nef -m contracts/proxygetter/proxygetter.manifest.json -c contracts/proxygetter/proxygetter.yml  
neo-go contract deploy -i contracts/proxygetter/proxygetter.nef -m contracts/proxygetter/proxygetter.manifest.json -r http://localhost:30333 -w wallets/wallet1.json

c91321703962769b66394dce84a3435cfdc78ed9f78dcf75f62efd68fb58551e
Contract: b168b196b0df31a8cea56204b09adeb8eec5d7c7
neo-go contract invokefunction -r http://localhost:30333 -w wallets/wallet1.json b168b196b0df31a8cea56204b09adeb8eec5d7c7 setToTestsNumber 0


# Добавление контрактов в группы

