package main

type StorageUnit int64

const (
	KB StorageUnit = 1 << 10
	MB StorageUnit = 1 << 20
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)

func sizeInBytes(value int64, unit StorageUnit) int64 {
	return value * int64(unit)
}

type Endpoint string

const (
	oauthAuthURL  Endpoint = "oauth_auth_url"
	oauthTokenURL Endpoint = "oauth_token_url"
)
