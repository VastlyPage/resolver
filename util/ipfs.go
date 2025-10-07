package hlutil

func GetGatewayURL(cid string) string {
	// TODO: Mitigate this single point of failure
	return cid + ".ipfs.dweb.link"
}
