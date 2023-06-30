package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func LoadAwsCredential(credentialsFile, profile string) *credentials.Credentials {
	// Create a session object.
	session, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials(credentialsFile, profile),
	})
	if err != nil {
		panic(err)
	}

	return session.Config.Credentials
}
