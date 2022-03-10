package client

import "github.com/go-resty/resty/v2"

var (
	RestyClient = resty.New()

)


func init() {
	RestyClient.SetDebug(false)

	RestyClient.SetBaseURL("http://127.0.0.1:8000")

	// Headers for all request
	//RestyClient.SetHeader("ContentType", "application/x-www-form-urlencoded")
	RestyClient.SetHeader("Content-Type","application/json")

	RestyClient.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))

	//TODO cookie 需要之后设置
	RestyClient.SetCloseConnection(true)

}
