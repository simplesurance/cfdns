#!/usr/bin/env bash

set -eu -o pipefail

CLIENT_SECRET=cAHaOwc1j6dMvyua

curl \
	-X POST \
	-H "Content-Type: application/x-www-form-urlencoded" \
	-d "client_id=ing&client_secret=$CLIENT_SECRET&grant_type=client_credentials" \
	https://accounts.sb-apidraft.sisu.sh/token

#curl -v \
#	-H "application/x-www-form-urlencoded" \
#	-F "client_id=ing" \
#	-F "client_secret=b" \
#	https://accounts.sb-apidraft.sisu.sh/token
