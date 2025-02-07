# FrostFS All-in-One

This repository contains Dockerfile for FrostFS All-in-One image and helper
scripts to start container. All-in-One image contains binaries and config
files for:
- neo-go
- frostfs-ir
- frostfs-storage
- frostfs-cli
- frostfs-adm
- frostfs-s3-gw
- frostfs-s3-authmate
- frostfs-http-gw

Entrypoint script starts blockchain, inner ring, storage, s3, and http gateway
services and configures it in the initial start.

| Service                     | Port             |
|-----------------------------|------------------|
| neo-go RPC                  | 30333            |
| FrostFS Storage gRPC API    | 8080             |
| FrostFS Storage Control API | 16513            |
| FrostFS HTTP Gateway        | 8081             |
| FrostFS S3 Gateway          | 8084, 8184 (TLS) |

# Limitations

All-in-One image contains single storage node, so the only viable policy for
containers is `REP 1`.


# Prerequisites

- docker v20.10 or higher
- docker-compose v2.10 or higher
- make v3.82 or higher

> Makefile script runs docker-compose with `--wait` flag, which is introduced in
> docker-compose v2. You can use older version of docker-compose by executing it
> manually.

# Quick Start

Clone repository and start container with docker-compose.

``` sh
$ git clone https://git.frostfs.info/TrueCloudLab/frostfs-aio.git
$ cd frostfs-aio
$ make up tick.epoch
```

Initial start initializes the storage configuration. Its readiness is based
on a healthcheck done by `docker-compose`.

Container can be stopped with:

``` sh
$ make down
```

Data and the system configuration is stored in container's volume.
Next time container is started, it will take less time to initialize.

Before updating image version, reset `frostfs-aio` by clearing its local volume
before starting container. Update may contain incompatible changes, e.g. neo-go
database format update. This repository does not provide any migration tools.

```
$ make clean
```

Also, you may have to make sure the storage node is in the network map.

``` sh
$ docker exec -ti aio frostfs-cli netmap snapshot -c /config/cli-cfg-sn.yaml --rpc-endpoint 127.0.0.1:8080
Epoch: 45
Node 1: 022bb4041c50d607ff871dec7e4cd7778388e0ea6849d84ccbd9aa8f32e16a8131 ONLINE /dns4/localhost/tcp/8080
    Continent: Europe
    Country: Germany
    CountryCode: DE
    Deployed: Private
    Location: Falkenstein
    Price: 10
    SubDiv: Sachsen
    SubDivCode: SN
    UN-LOCODE: DE FKS
```

If you don't see the output like this, you can wait for the new Epoch to come
or force the starting of new epoch.

``` sh
$ make tick.epoch
Current epoch: 1, increase to 2.
Waiting for transactions to persist...
```

# Build images

## Standard image

Build frostfs-aio image with this command.

``` sh
$ make image-aio
```

## Image with local binaries

Put all the needed pre-built binaries to the `frostfs-aio/bin/`:

- neo-go
- frostfs-adm
- frostfs-cli
- frostfs-ir
- frostfs-node
- frostfs-s3-gw
- frostfs-s3-authmate
- frostfs-http-gw

Make sure they all have the `x` right for the execution enabled.
Build frostfs-aio image using pre-built local binaries for all the services.

``` sh
$ make image-aio-local
```

## Image with custom binaries

This target is useful when one needs to change only a few binaries and keep the others "as is".

Put the needed binaries to `frostfs-aio/bin/`. Make sure they all have the `x` right for the
execution enabled. All the binaries will be copied to image.

``` sh
$ make image-aio-custom
```

# Simple storage


## S3 interface

### Setup S3

As soon as the storage node is in the network map (see above) you can generate S3
credentials:

``` sh
$ make s3cred
{
  "access_key_id": "EXArWh8x1zeHG3851s1RtoCo7dowxF6rhLGA15nbMffT0AKRSjJ5fmcqf3Ht2VCAkfmPQUVARghRB77xHCA1BoN2p",
  "secret_access_key": "d70c1dba83f0f90bb231f06f1ce0e0dfbcfb122f4b4345a3c18d3869c359b79f",
  "owner_private_key": "140947599afd9ca89af4b358c3176eb046e554d942a0dc99a8e06f3e43c8f4ad",
  "wallet_public_key": "0324e76288fcb900100d01802a14ef977cca45ad073561230446df14b344c858b6",
  "container_id": "EXArWh8x1zeHG3851s1RtoCo7dowxF6rhLGA15nbMffT"
}                   
```

Now let's configure an S3 client (AWS CLI will be used as example):

``` sh
$ aws configure
AWS Access Key ID []: EXArWh8x1zeHG3851s1RtoCo7dowxF6rhLGA15nbMffT0AKRSjJ5fmcqf3Ht2VCAkfmPQUVARghRB77xHCA1BoN2p
AWS Secret Access Key []: d70c1dba83f0f90bb231f06f1ce0e0dfbcfb122f4b4345a3c18d3869c359b79f
Default region name []: us-east-1
Default output format []: json
```

### Create a container

``` sh
$ aws s3api --endpoint http://localhost:8084 create-bucket --bucket koty --acl public-read-write
```

### Put an object

``` sh
$ aws s3api --endpoint http://localhost:8084 put-object --bucket koty --key kot --body cat.jpg
{                                                                                                                                                                                                                                                                               
    "ETag": "8677919550a90ff7106584285f25a70ac9e7aa38bdb4ed17d34bbfb366fd71b7"                                                                                                                                                                                                  
} 
```

### List objects

``` sh
$ aws s3api --endpoint http://localhost:8084 list-objects --bucket koty
{
    "Contents": [
        {
            "Key": "kot",
            "LastModified": "2023-02-17T15:01:52+00:00",
            "ETag": "8677919550a90ff7106584285f25a70ac9e7aa38bdb4ed17d34bbfb366fd71b7",
            "Size": 174512,
            "Owner": {
                "DisplayName": "NWeByJPgNC97F83hTUnSbnZSBKaFvk5HNw",
                "ID": "NWeByJPgNC97F83hTUnSbnZSBKaFvk5HNw"
            }
        }
    ]
}
```

### S3 credentials from custom wallets

Credentials made by `make s3cred` command are based on the private key from 
`s3-gw/user-wallet.json` file. If you need to create credential for different
users, use wallets from `wallets` dir.

```sh
$ make s3cred-custom wallet=wallet2.json
{
  "access_key_id": "jHhL5B33o16R4jQsb8wm9A3RRdS6KrTB5N4bja9Jys904W7xXFNKqem2ACvTRWRYJsZMCUikYFSokN7pPJziWyDi",
  "secret_access_key": "21bb64fafa32c82417fd8b97ac56cc8a085998a3852632d52fe7042453daa440",
  "owner_private_key": "10f6f9d7a47bb0bf68363ad8a99fe69f1493f8b6e1665b3e4e83feb2d5c7ee39",
  "wallet_public_key": "03e38759973a6bb722baabc2dd84036a39f0b2f53d32fec45a4dacde8a50fe4b70",
  "container_id": "jHhL5B33o16R4jQsb8wm9A3RRdS6KrTB5N4bja9Jys9"
}
```

To get credentials from custom wallet, place it in `wallets` dir before start.
Make sure that wallet account has empty password.

### Making a custom wallet

Command line interface of `neo-go` is used to create new wallets. Both `neo-go`
from the image or [downloaded binary](https://github.com/nspcc-dev/neo-go/releases)
can be used. We stick to the one in the image in this example.

Invoke a shell of the running container:
```bash
$ docker exec -it aio bash
```

Create a new wallet:

```bash
$ neo-go wallet init -a --wallet /wallets/new-wallet.json
```

- key `-a` means the new wallet must have an account.
- key `--wallet /wallets/new-wallet.json` sets the outputh path for the new wallet.
The image keeps wallets at `/wallets` for simplicity.

Then `neo-go` asks for the account name and its password using interactive
mode. Leave the password empty.

```bash
Enter the name of the account > Some Name
Enter new password >
Confirm password >
``` 

If everything is rigth, the stdout will show a pretty-printed contents of the wallet:
```bash
{
        "version": "1.0",
        "accounts": [
                {
                        "address": "NapWfRshCv9T3UyGxSfPr6NKiTFa5sVwQx",
                        "key": "6PYKKPtBRuKvTqjAuUZk5hVDgG5aZ3HQvp1hSQg5jDyUF9JaP5Jcy3kWiV",
                        "label": "Some Name",
                        "contract": {
                                "script": "DCECGXJggQTMyMhYS9YtWzTP+3FbiGLn2h7za1zSNgMDkIxBVuezJw==",
                                "parameters": [
                                        {
                                                "name": "parameter0",
                                                "type": "Signature"
                                        }
                                ],
                                "deployed": false
                        },
                        "lock": false,
                        "isDefault": false
                }
        ],
        "scrypt": {
                "n": 16384,
                "r": 8,
                "p": 8
        },
        "extra": {
                "Tokens": null
        }
 }

wallet successfully created, file location is /wallets/new-wallet.json
```

You can use `make wallet WALLET=wallets/new-wallet.json`
to execute the steps listed above a little faster.


### Transfering GAS tokens to wallet

To add GAS to a new wallet use `make refill`

```console
$ make refill GAS=12.3 WALLET=wallets/new-wallet.json
Waiting for transactions to persist...
Balance for wallets/new-wallet.json
Account NgB6jnyVsMg6E3gQQLXoccZ3JBMLsgiut3
GAS: GasToken (d2a4cff31913016155e38e474a2c06d08be276cf)
        Amount : 12.3
        Updated: 25069
```

### Checking wallet balance

```console
$ make balance WALLET=wallets/new-wallet.json
Balance for wallets/new-wallet.json
Account NgB6jnyVsMg6E3gQQLXoccZ3JBMLsgiut3
GAS: GasToken (d2a4cff31913016155e38e474a2c06d08be276cf)
        Amount : 12.3
        Updated: 25069
```


## frostfs-cli interface

### Native credentials

To use wallet file, it should be register in FrostFS ID contract and have  
policies to perform FrostFS operations.

Use `make cred` to register `s3-gw/user-wallet.json`. For custom wallets from
`wallet` dir, use `make cred-custom wallet=wallet1.json`.

These commands are similar to `make s3cred` and `make s3cred-custom`.

### Create container with frostfs-cli

``` sh
$ frostfs-cli -r localhost:8080 -w s3-gw/user-wallet.json \
            --address NWeByJPgNC97F83hTUnSbnZSBKaFvk5HNw \
            container create \
            --policy "REP 1" --await
container ID: GfWw35kHds7gKWmSvW7Zi4U39K7NMLK8EfXBQ5FPJA46
awaiting...
container has been persisted on sidechain
```

### Put an object with frostfs-cli

``` sh
$ frostfs-cli -r localhost:8080 -w s3-gw/user-wallet.json \
            --address NWeByJPgNC97F83hTUnSbnZSBKaFvk5HNw \
            object put \
            --cid GfWw35kHds7gKWmSvW7Zi4U39K7NMLK8EfXBQ5FPJA46 \
            --file cat.jpg
[cat.jpg] Object successfully stored
  OID: HByVC9A34i22BnzW3n83z9vEMxuYZoC7nNu11ZvGeCTe
  CID: GfWw35kHds7gKWmSvW7Zi4U39K7NMLK8EfXBQ5FPJA46
```

### Get and object with frostfs-cli

``` sh
$ frostfs-cli -r localhost:8080 -w s3-gw/user-wallet.json \
            --address NWeByJPgNC97F83hTUnSbnZSBKaFvk5HNw \
            object get \
            --cid GfWw35kHds7gKWmSvW7Zi4U39K7NMLK8EfXBQ5FPJA46 \
            --oid HByVC9A34i22BnzW3n83z9vEMxuYZoC7nNu11ZvGeCTe > new_cat.jpg
```

## Viewing logs

There are two targets for viewing logs.

You can view all the logs by replacing the desired service name:

``` sh
make logs/frostfs-node
```

You can also make a continuous stream of logs by replacing the desired service name:

``` sh
make stream-logs/frostfs-node
```
