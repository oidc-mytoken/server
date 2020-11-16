package main

import (
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	configurationEndpoint "github.com/zachmann/mytoken/internal/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/jws"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
	"github.com/zachmann/mytoken/internal/server"
	loggerUtils "github.com/zachmann/mytoken/internal/utils/logger"
)

func main() {
	config.Load()
	loggerUtils.Init()
	server.Init()
	configurationEndpoint.Init()
	authcode.Init()
	if err := db.Connect(); err != nil {
		panic(err)
	}
	jws.LoadKey()

	server.Start()

	//st := supertoken.NewSuperToken("ggggggggggggggggggggggg", "https://oidc.issuer.data.kit.edu/long/url/i/test/how/long/the/token/will/get", restrictions.Restrictions{{UsagesAT: 1, Audiences: []string{"https://some.service.data.kit.edu/"}, Scope: "compute store read write"}}, capabilities.Capabilities{"AT"})
	//jwt, _ := st.ToJWT()
	//fmt.Printf("%s\n", jwt)

	// st, err := supertoken.NewSuperTokenEntry("testToken", "gabriel", "wlcg", restrictions.Restrictions{
	// 	{
	// 		NotBefore: 1599939600,
	// 		ExpiresAt: 1599948600,
	// 		IPs:       []string{"192.168.0.31"},
	// 		UsagesAT:  10,
	// 	}, {
	// 		NotBefore: 1599939600,
	// 		ExpiresAt: 1599940600,
	// 		Scope:     "storage",
	// 	},
	// }, capabilities.Capabilities{"AT", "create_super_token"})
	// if err != nil {
	// 	panic(err)
	// }
	// jwt, err := st.Token.ToJWT()
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(jwt)
	// 	fmt.Println()
	// 	st, err := supertoken.ParseJWT(jwt)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	} else {
	// 		fmt.Printf("%+v\n", st)
	// 	}
	// }
	// fmt.Println()
}
