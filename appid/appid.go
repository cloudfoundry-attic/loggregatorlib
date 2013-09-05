package appid

import (
	"code.google.com/p/gogoprotobuf/proto"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"net/url"
)

func FromUrl(u *url.URL) string {
	appId := u.Query().Get("app")
	return appId
}

func FromLogMessage(data []byte) (appId string, drainUrls []string, err error) {
	receivedMessage := &logmessage.LogMessage{}
	err = proto.Unmarshal(data, receivedMessage)
	if err != nil {
		err = errors.New(fmt.Sprintf("Log message could not be unmarshaled. Dropping it... Error: %v. Data: %v", err, data))
		return "", make([]string, 0), err
	}

	return *receivedMessage.AppId, receivedMessage.GetDrainUrls(), nil
}
