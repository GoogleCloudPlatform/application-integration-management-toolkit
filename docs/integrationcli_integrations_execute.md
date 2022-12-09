## integrationcli integrations execute

Execute an integration

### Synopsis

execute an integration

```
integrationcli integrations execute [flags]
```

### Options

```
  -f, --file string   Integration payload JSON file path. For the payload structure, visit docs at https://cloud.google.com/application-integration/docs/reference/rest/v1/projects.locations.integrations/execute#request-body
  -h, --help          help for execute
  -n, --name string   Integration flow name
```

### Options inherited from parent commands

```
  -a, --account string       Path Service Account private key in JSON
      --apigee-integration   Use Apigee Integration; default is false (Application Integration)
      --disable-check        Disable check for newer versions
      --no-output            Disable printing API responses from the control plane
  -p, --proj string          Integration GCP Project name
  -r, --reg string           Integration region name
  -t, --token string         Google OAuth Token
```

### SEE ALSO

* [integrationcli integrations](integrationcli_integrations.md)	 - Manage integrations in a GCP project

###### Auto generated by spf13/cobra on 29-Nov-2022