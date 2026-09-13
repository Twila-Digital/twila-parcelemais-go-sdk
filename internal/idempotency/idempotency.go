package idempotency

import "strings"

var mutablePathsRequiringKey = []string{
	"v1/order/start-cdc-sale",
	"v1/order/invoice",
	"v1/order",
	"v1/webhooks",
}

func RequiresKey(method, path string) bool {
	if strings.ToUpper(method) != "POST" {
		return false
	}
	for _, p := range mutablePathsRequiringKey {
		if strings.HasSuffix(path, p) {
			return true
		}
	}
	return false
}

func IsRetrySafe(method, path string, hasKey bool) bool {
	upper := strings.ToUpper(method)

	switch upper {
	case "GET", "HEAD", "OPTIONS":
		return true
	case "PUT", "DELETE":
		return strings.Contains(path, "v1/webhooks/")
	case "POST":
		return RequiresKey(method, path) && hasKey
	default:
		return false
	}
}
