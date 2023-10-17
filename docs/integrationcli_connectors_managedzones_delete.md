## integrationcli connectors managedzones delete

Delete a managedzone configuration

### Synopsis

Delete a managedzone configuration

```
integrationcli connectors managedzones delete [flags]
```

### Options

```
  -h, --help          help for delete
  -n, --name string   The name of the managedzone
```

### Options inherited from parent commands

```
  -a, --account string   Path Service Account private key in JSON
      --api api          Sets the control plane API. Must be one of prod, staging or autopush; default is prod
      --disable-check    Disable check for newer versions
      --metadata-token   Metadata OAuth2 access token
      --no-output        Disable printing all statements to stdout
      --print-output     Control printing of info log statements (default true)
  -p, --proj string      Integration GCP Project name
  -r, --reg string       Integration region name
  -t, --token string     Google OAuth Token
      --verbose          Enable verbose output from integrationcli
```

### SEE ALSO

* [integrationcli connectors managedzones](integrationcli_connectors_managedzones.md)	 - Manage DNS Peering with Integration Connectors

###### Auto generated by spf13/cobra on 4-Oct-2023