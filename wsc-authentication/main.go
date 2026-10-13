package main

import (
	"fmt"
	"os"
	"strings"

	nex "github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-common-go/authentication"
)

var nexServer *nex.Server

func main() {
	secureLocation := strings.TrimSpace(os.Getenv("SECURE_SERVER_LOCATION"))
	if secureLocation == "" {
		fmt.Fprintln(os.Stderr, "SECURE_SERVER_LOCATION is required: set the public host/IP advertised to Wii U clients")
		os.Exit(2)
	}
	if strings.TrimSpace(os.Getenv("KERBEROS_PASSWORD")) == "" {
		fmt.Fprintln(os.Stderr, "KERBEROS_PASSWORD is required")
		os.Exit(2)
	}
	if err := connectMongo(); err != nil {
		fmt.Fprintf(os.Stderr, "NEX account database startup failed: %v\n", err)
		os.Exit(1)
	}

	startInternalNEXTokenAPI()

	nexServer = nex.NewServer()
	nexServer.SetPRUDPVersion(1)
	nexServer.SetPRUDPProtocolMinorVersion(3)
	nexServer.SetDefaultNEXVersion(&nex.NEXVersion{
		Major: 3,
		Minor: 4,
		Patch: 0,
	})
	nexServer.SetKerberosPassword(os.Getenv("KERBEROS_PASSWORD"))
	nexServer.SetAccessKey("4d324052")

	nexServer.On("Data", func(packet *nex.PacketV1) {
		request := packet.RMCRequest()
		fmt.Printf("==WSC Auth== proto=%#v method=%#v params_len=%d\n",
			request.ProtocolID(), request.MethodID(), len(request.Parameters()))
	})

	authenticationProtocol := authentication.NewCommonAuthenticationProtocol(nexServer)

	securePort := strings.TrimSpace(os.Getenv("SECURE_SERVER_PORT"))
	if securePort == "" {
		securePort = "60015"
	}
	secureStationURL := nex.NewStationURL("")
	secureStationURL.SetScheme("prudps")
	secureStationURL.SetAddress(secureLocation)
	secureStationURL.SetPort(securePort)
	secureStationURL.SetCID("1")
	secureStationURL.SetPID("2")
	secureStationURL.SetSID("1")
	secureStationURL.SetStream("10")
	secureStationURL.SetType("2")

	authenticationProtocol.SetSecureStationURL(secureStationURL)
	authenticationProtocol.SetBuildName("Community WSC")
	authenticationProtocol.SetPasswordFromPIDFunction(passwordFromPID)

	nexServer.Listen(":60014")
}
