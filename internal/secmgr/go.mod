module secmgr

require (
	internal/apiclient v1.0.0
	internal/client v1.0.0
	internal/clilog v1.0.0
	internal/cloudkms v1.0.0
	internal/cmd v1.0.0
	internal/secmgr v1.0.0
)

replace internal/apiclient => ../../internal/apiclient

replace internal/clilog => ../../internal/clilog

replace internal/cloudkms => ../../internal/cloudkms

replace internal/secmgr => ../../internal/secmgr

replace internal/client => ../../internal/client

replace internal/cmd => ../../internal/cmd

go 1.23.2
