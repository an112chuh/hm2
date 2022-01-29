package constants

type CookiesSettings struct {
	AuthKeyOne    []byte
	EncryptionOne []byte
}

var Cookies CookiesSettings

func SetCookies() {
	Cookies.AuthKeyOne = []byte("frhueiuhefhuryfgdbshtyrueivngkdjeyrhgubndpwlmnazxfhryetvbqwdfhty")
	Cookies.EncryptionOne = []byte("yitoeqptohlcurmwncbakfiyvdgtlpmz")
}
