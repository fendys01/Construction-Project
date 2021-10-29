package agora

import (
	"time"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
)

type Contract struct {
	AppID   string
	AppCert string
}

// Use RtcTokenBuilder to generate an RTC token.
func (c Contract) GenerateRtcToken(initUID uint32, channelName string, role rtctokenbuilder.Role) (string, error) {
	appID := c.AppID
	appCertificate := c.AppCert

	// Number of seconds after which the token expires.
	// For demonstration purposes the expiry time is set to 40 seconds. This shows you the automatic token renew actions of the client.
	expireTimeInSeconds := uint32(40)

	// Get current timestamp.
	currentTimestamp := uint32(time.Now().UTC().Unix())

	// Timestamp when the token expires.
	expireTimestamp := currentTimestamp + expireTimeInSeconds

	return rtctokenbuilder.BuildTokenWithUID(appID, appCertificate, channelName, initUID, role, expireTimestamp)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Printf("Token with uid: %s\n", result)
	// 	fmt.Printf("uid is %d\n", initUID)
	// 	fmt.Printf("ChannelName is %s\n", channelName)
	// 	fmt.Printf("Role is %d\n", role)
	// }
}
