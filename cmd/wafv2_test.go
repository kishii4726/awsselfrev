package cmd

import (
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWAFV2Client struct {
	mock.Mock
}

func (m *MockWAFV2Client) ListWebACLs(ctx context.Context, params *wafv2.ListWebACLsInput, optFns ...func(*wafv2.Options)) (*wafv2.ListWebACLsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*wafv2.ListWebACLsOutput), args.Error(1)
}

func (m *MockWAFV2Client) GetLoggingConfiguration(ctx context.Context, params *wafv2.GetLoggingConfigurationInput, optFns ...func(*wafv2.Options)) (*wafv2.GetLoggingConfigurationOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wafv2.GetLoggingConfigurationOutput), args.Error(1)
}

func TestCheckWAFV2Configurations(t *testing.T) {
	regClient := new(MockWAFV2Client)
	cfClient := new(MockWAFV2Client)

	regionalACLs := []types.WebACLSummary{
		{Name: aws.String("reg-acl"), ARN: aws.String("arn:reg")},
	}
	cfACLs := []types.WebACLSummary{
		{Name: aws.String("cf-acl"), ARN: aws.String("arn:cf")},
	}

	regClient.On("ListWebACLs", mock.Anything, mock.MatchedBy(func(p *wafv2.ListWebACLsInput) bool {
		return p.Scope == types.ScopeRegional
	}), mock.Anything).Return(&wafv2.ListWebACLsOutput{WebACLs: regionalACLs}, nil)

	cfClient.On("ListWebACLs", mock.Anything, mock.MatchedBy(func(p *wafv2.ListWebACLsInput) bool {
		return p.Scope == types.ScopeCloudfront
	}), mock.Anything).Return(&wafv2.ListWebACLsOutput{WebACLs: cfACLs}, nil)

	// Reg: Fail logging, CF: Pass logging
	regClient.On("GetLoggingConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*wafv2.GetLoggingConfigurationOutput)(nil), &types.WAFNonexistentItemException{})
	cfClient.On("GetLoggingConfiguration", mock.Anything, mock.Anything, mock.Anything).Return(&wafv2.GetLoggingConfigurationOutput{
		LoggingConfiguration: &types.LoggingConfiguration{},
	}, nil)

	tbl := table.SetTable()
	rules := config.RulesConfig{
		Rules: map[string]config.Rule{
			"wafv2-logging-enabled": {Service: "WAFV2", Level: "Warning", Issue: "Logging is not enabled"},
		},
	}

	checkWAFV2Configurations(regClient, cfClient, tbl, rules)

	assert.Equal(t, 2, tbl.NumLines())
}
