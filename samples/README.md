# integrationcli command Samples

The following table contains some examples of integrationcli.

Set up integrationcli with preferences: `integrationcli prefs set -p $project -r $region`

| Operations | Command |
|---|---|
| integrations | `integrationcli integrations create -n $name -f samples/sample.json -u $userLabel --default-token`|
| integrations | `integrationcli integrations create -n $name -f samples/sample.json -o samples/sample_overrides.json --default-token`|
| integrations | `integrationcli integrations create -n $name -f samples/sample.json --publish=true --default-token`|
| integrations | `integrationcli integrations create -n $name -f samples/sample.json --basic=true --default-token`|
| integrations | `integrationcli integrations versions list -n $integration --basic=true --default-token`|
| integrations | `integrationcli integrations versions list -n $integration --basic=true --filter=state=ACTIVE --default-token`|
| integrations | `integrationcli integrations scaffold -n $integration -s $snapshot -f . --env=dev --default-token`|
| integrations | `integrationcli integrations scaffold -n $integration -s $snapshot -f . --skip-connectors=true --default-token`|
| integrations | `integrationcli integrations scaffold -n $integration -s $snapshot -f . --cloud-build=true --default-token`|
| integrations | `integrationcli integrations scaffold -n $integration -s $snapshot -f . --cloud-deploy=true --default-token`|
| integrations | `integrationcli integrations apply -f . --wait=true --default-token`|
| integrations | `integrationcli integrations apply -f . --env=dev --default-token`|
| integrations | `integrationcli integrations apply -f . --grant-permission=true --default-token`|
| integrations | `integrationcli integrations apply -f . --skip-connectors=true --default-token`|
| integrations | `integrationcli integrations versions publish -n $name --default-token`|
| integrations | `integrationcli integrations versions publish -n $name -s $snapshot --default-token`|
| integrations | `integrationcli integrations versions unpublish -n $name --default-token`|
| integrations | `integrationcli integrations versions unpublish -n $name -u $userLabel --default-token`|
| integrations | `integrationcli integrations apply -f . --env=dev --tests-folder=./test-configs --default-token`|
| authconfigs | `integrationcli authconfigs create -f samples/ac_username.json`|
| authconfigs | `integrationcli authconfigs create -f samples/ac_oidc.json`|
| authconfigs | `integrationcli authconfigs create -f samples/ac_authtoken.json`|
| authconfigs | `integrationcli authconfigs create -e samples/b64encoded_ac.txt -k locations/$region/keyRings/$key/cryptoKeys/$cryptokey`|
| connectors | `integrationcli connectors create -n $name -g=true -f samples/pub_sub_connection.json -sa=connectors --wait=true --default-token`|
| GitHub Actions | See samples [here](../samples/workflows) |



NOTE: This file is auto-generated during a release. Do not modify.