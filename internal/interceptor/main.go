package interceptor

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"theori.io/aws-terraform-boost/internal/types"
)

func AWSInterceptors() map[string]func(url.Values, *types.ActionCache, *types.Account, *aws.Config) []byte {
	actionHandlers := make(map[string]func(url.Values, *types.ActionCache, *types.Account, *aws.Config) []byte)
	actionHandlers["DescribeSecurityGroupRules"] = DescribeSecurityGroupRules
	return actionHandlers
}
