# vault-plugin-secrets-mailgun
[![Build Status](https://travis-ci.org/usr42/vault-plugin-secrets-mailgun.svg?branch=master)](https://travis-ci.org/usr42/vault-plugin-secrets-mailgun)

**This plugin is still in early development, not at all feature complete
or usable. Don't use it in production.**

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
