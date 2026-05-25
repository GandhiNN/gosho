package sso

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
)

type AccountInfo struct {
	Name string
	ID   string
}

func (a AccountInfo) String() string {
	return fmt.Sprintf("%s - %s", a.Name, a.ID)
}

func ListAccounts(
	ctx context.Context,
	client *sso.Client,
	accessToken string,
) ([]AccountInfo, error) {
	var accounts []AccountInfo
	var nextToken *string
	for {
		resp, err := client.ListAccounts(ctx, &sso.ListAccountsInput{
			AccessToken: &accessToken,
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("list accounts: %w", err)
		}
		for _, a := range resp.AccountList {
			accounts = append(accounts, AccountInfo{
				Name: aws.ToString(a.AccountName),
				ID:   aws.ToString(a.AccountId),
			})
		}
		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return accounts, nil
}

func ListRoles(
	ctx context.Context,
	client *sso.Client,
	accessToken, accountID string,
) ([]string, error) {
	var roles []string
	var nextToken *string
	for {
		resp, err := client.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
			AccessToken: &accessToken,
			AccountId:   &accountID,
			MaxResults:  aws.Int32(50),
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("list roles: %w", err)
		}
		for _, r := range resp.RoleList {
			roles = append(roles, aws.ToString(r.RoleName))
		}
		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}
	return roles, nil
}

type RoleCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      int64
}

func GetRoleCredentials(
	ctx context.Context,
	client *sso.Client,
	accessToken, accountID, role string,
) (*RoleCredentials, error) {
	resp, err := client.GetRoleCredentials(ctx, &sso.GetRoleCredentialsInput{
		AccessToken: &accessToken,
		AccountId:   &accountID,
		RoleName:    &role,
	})
	if err != nil {
		return nil, fmt.Errorf("get role credentials: %w", err)
	}
	return &RoleCredentials{
		AccessKeyID:     aws.ToString(resp.RoleCredentials.AccessKeyId),
		SecretAccessKey: aws.ToString(resp.RoleCredentials.SecretAccessKey),
		SessionToken:    aws.ToString(resp.RoleCredentials.SessionToken),
		Expiration:      resp.RoleCredentials.Expiration,
	}, nil
}
