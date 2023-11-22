#!/usr/bin/env bash

export TEST_CF_ZONE_NAME=simplesurance.top

TEST_CF_APITOKEN=$(vault kv get --field=api_token covert/apps/xsell/nats-dns-manager-service/sisu-sandbox-eu/cloudflare)
export TEST_CF_APITOKEN
