package types

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"theori.io/aws-terraform-boost/internal/utils"
)

type Account struct {
	session *session.Session
}

func NewAccount(credentialsFile, profile string) *Account {
	session, _ := session.NewSession(&aws.Config{
		Credentials: utils.LoadAwsCredential(credentialsFile, profile),
	})

	return &Account{session}
}

func (account *Account) NewEC2(cfg *aws.Config) *ec2.EC2 {
	return ec2.New(account.session, cfg)
}
