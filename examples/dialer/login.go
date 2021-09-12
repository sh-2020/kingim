package dialer

import (
	"kingim"
	"kingim/websocket"
	"kingim/wire/token"
)

func Login(wsurl, account string, appSecrets ...string) (kingim.Client, error) {
	cli := websocket.NewClient(account, "unittest", websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
		secret = appSecrets[0]
	}
	cli.SetDialer(&ClientDialer{
		AppSecret: secret,
	})
	err := cli.Connect(wsurl)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
