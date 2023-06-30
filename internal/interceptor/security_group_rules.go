package interceptor

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"theori.io/aws-terraform-boost/internal/types"
	"theori.io/aws-terraform-boost/internal/utils"
)

func DescribeSecurityGroupRules(q url.Values, cache *types.ActionCache, account *types.Account, config *aws.Config) []byte {
	// Widen query parameters to fetch all entries
	id := q.Get("SecurityGroupRuleId.1")
	q.Del("SecurityGroupRuleId.1")

	key := q.Encode()
	resp := cache.Entry(key, func() interface{} {
		var out []*ec2.SecurityGroupRule
		account.NewEC2(config).DescribeSecurityGroupRulesPages(
			&ec2.DescribeSecurityGroupRulesInput{},
			func(page *ec2.DescribeSecurityGroupRulesOutput, a bool) bool {
				out = append(out, page.SecurityGroupRules...)
				return true
			},
		)
		return out
	})

	return filter(resp, id)
}

func filter(resp *types.CacheEntry, id string) []byte {
	f := resp.Body
	new_items := make([]*ec2.SecurityGroupRule, 0)

	for _, item := range f.([]*ec2.SecurityGroupRule) {
		if *item.SecurityGroupRuleId == id {
			new_items = append(new_items, item)
		}
	}

	res := ec2.DescribeSecurityGroupRulesOutput{SecurityGroupRules: new_items, NextToken: nil}
	return []byte(utils.StringifyXML(res, "DescribeSecurityGroupRulesResponse"))
}
