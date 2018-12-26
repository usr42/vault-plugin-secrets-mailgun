#!/bin/sh

LONG_NAME=vault-plugin-secrets-mailgun
SHORT_NAME=mailgun

docker stop $LONG_NAME
env GOOS=linux GOARCH=amd64 go build
export SHA256=$(shasum -a 256 "$LONG_NAME" | cut -d' ' -f1)
mv $LONG_NAME docker/
cd docker
docker build -t $LONG_NAME .
rm $LONG_NAME
docker run --cap-add=IPC_LOCK --rm -e 'VAULT_DEV_ROOT_TOKEN_ID=myroot' -p8200:8200 -d --name $LONG_NAME $LONG_NAME
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN="myroot"
sleep 1
vault write sys/plugins/catalog/$SHORT_NAME \
    sha_256="${SHA256}" \
    command="$LONG_NAME"
vault secrets enable \
    -path="$SHORT_NAME" \
    -plugin-name="$SHORT_NAME" \
    plugin
echo
echo "To communicate with vault souce ./env:"
echo ". ./env"
