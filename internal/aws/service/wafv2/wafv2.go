package service

import (
	"awsselfrev/internal/aws/api"
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

type WebACLInfo struct {
	Name string
	ARN  string
}

func ListWebACLs(client api.WAFV2Client, scope types.Scope) []WebACLInfo {
	var webACLs []WebACLInfo
	resp, err := client.ListWebACLs(context.TODO(), &wafv2.ListWebACLsInput{
		Scope: scope,
	})
	if err != nil {
		log.Printf("Warning: Failed to list WAF v2 Web ACLs for scope %s: %v", scope, err)
		return webACLs
	}

	for _, acl := range resp.WebACLs {
		webACLs = append(webACLs, WebACLInfo{
			Name: *acl.Name,
			ARN:  *acl.ARN,
		})
	}
	return webACLs
}

func IsWAFV2LoggingEnabled(client api.WAFV2Client, resourceArn string) bool {
	_, err := client.GetLoggingConfiguration(context.TODO(), &wafv2.GetLoggingConfigurationInput{
		ResourceArn: aws.String(resourceArn),
	})
	if err != nil {
		var nfe *types.WAFNonexistentItemException
		if errors.As(err, &nfe) {
			return false
		}
		log.Printf("Warning: Failed to get WAF v2 logging configuration for %s: %v", resourceArn, err)
		return false
	}
	return true
}
