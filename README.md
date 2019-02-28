# vault-plugin-secrets-mailgun
[![Build Status](https://travis-ci.org/usr42/vault-plugin-secrets-mailgun.svg?branch=master)](https://travis-ci.org/usr42/vault-plugin-secrets-mailgun)
[![Go Report Card](https://goreportcard.com/badge/github.com/usr42/vault-plugin-secrets-mailgun)](https://goreportcard.com/report/github.com/usr42/vault-plugin-secrets-mailgun)

This is a backend plugin to be used with
[Hashicorp Vault](https://www.github.com/hashicorp/vault).
It creates dynamic [mailgun](https://www.mailgun.com/) SMTP credentials
for a given domain.


## Quick Links
- [Vault Website](https://www.vaultproject.io)
- [Vault Github Project](https://www.github.com/hashicorp/vault)
- [mailgun API Reference](https://documentation.mailgun.com/en/latest/api_reference.html)

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already
installed Vault and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to
[get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on
[Vault plugins](https://www.vaultproject.io/docs/internals/plugins.html).

### Setup

The setup guide assumes you have a Vault server already running,
unsealed, and authenticated. The vault configuration has a
`plugin_directory` and a `api_addr` set (see 
[vault.hcl](docker/vault.hcl)).

1. Clone this repository and compile it:
    ```sh
    git clone https://github.com/usr42/vault-plugin-secrets-mailgun

    cd vault-plugin-secrets-mailgun

    go build
    ```

2. Move the compiled plugin into Vault's configured `plugin_directory`
(here: `/vault/plugins/`):

    ```sh
    mv vault-plugin-secrets-mailgun /vault/plugins/vault-plugin-secrets-mailgun
    ```

3. Calculate the SHA256 of the plugin and register it in Vault's plugin
catalog.

     ```sh
     export SHA256=$(shasum -a 256 "/vault/plugins/vault-plugin-secrets-mailgun" | cut -d' ' -f1)

     vault write sys/plugins/catalog/mailgun \
         sha_256="${SHA256}" \
         command="vault-plugin-secrets-mailgun"
     ```

4. Mount the plugin:

    ```sh
    vault secrets enable \
        -path="mailgun" \
        -plugin-name="mailgun" \
        plugin
    ```

### Configure Mailgun

Every plugin instance (here we used the path `mailgun/`) has to be configured
with one of your Mailgun domains and your Mailgun API key. Your domains can be
found on the [Mailgun Domains page](https://app.mailgun.com/app/domains/).
After clicking on one of the domains you see your API Key.

To configure the plugin with the API key `yourapikey` for the domain
`example.com` run:
```sh
$ vault write mailgun/config api_key=yourapikey domain=example.com
Success! Data written to: mailgun/config
```
The domain and the API key are mandatory and have to be valid or the write
command will fail.

Additionally you can configure `ttl` (the default TTL for each credential) and
`max_ttl` (the maximum TTL, even with refresh, for each credential)

If the credential is not refreshed within the TTL it will automatically be
revoked.

### Usage

After the secrets engine is configured it can generate credentials.
Generate a new credential by reading from the `/credentials` endpoint:

```sh
$ vault read mailgun/credentials
Key                Value
---                -----
lease_id           mailgun/credentials/3OYD4hLzGPrKRukI9RKIrFkc
lease_duration     768h
lease_renewable    true
password           5dqbj6YcALoQ3weD9ZDlszJWIpNueOt0
username           vault.1yrqc@example.com
```

The credentials can be refreshed or revoked like described in the
[Vault documentation - Lease, Renew, and Revoke](https://www.vaultproject.io/docs/concepts/lease.html)
