package db

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var (
	Con *sql.DB
)

type dbUser struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"`
}

func getDSNAWS() string {
	secretName := "mplinkstersdb"
	region := "us-west-1"
	u := dbUser{}

	if getEnv("MPDBPASSWORD", "") == "" {
		//Create a Secrets Manager client
		log.Printf("Creating AWS session")
		sess, err := session.NewSession()
		if err != nil {
			// Handle session creation error
			fmt.Println(err.Error())
			return ""
		}
		log.Printf("Creating SecretManager object")
		svc := secretsmanager.New(sess,
			aws.NewConfig().WithRegion(region))
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
		}

		// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
		// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html

		result, err := svc.GetSecretValue(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case secretsmanager.ErrCodeDecryptionFailure:
					// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
					fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

				case secretsmanager.ErrCodeInternalServiceError:
					// An error occurred on the server side.
					fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

				case secretsmanager.ErrCodeInvalidParameterException:
					// You provided an invalid value for a parameter.
					fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

				case secretsmanager.ErrCodeInvalidRequestException:
					// You provided a parameter value that is not valid for the current state of the resource.
					fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

				case secretsmanager.ErrCodeResourceNotFoundException:
					// We can't find the resource that you asked for.
					fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return ""
		}

		// Decrypts secret using the associated KMS CMK.
		// Depending on whether the secret is a string or binary, one of these fields will be populated.
		secretString := *result.SecretString

		// Your code goes here.
		err = json.Unmarshal([]byte(secretString), &u)
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
	} else {
		u.Password = getEnv("MPDBPASSWORD", "")
		u.Username = "admin"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", u.Username, u.Password, u.Host, u.Port, u.DbInstanceIdentifier)
}

func getDSNEnv() string {
	username := getEnv("MPDBUSER", "root")
	password := getEnv("MPDBPASSWORD", "")
	host := getEnv("MPDBHOST", "localhost")
	port := getEnv("MPDBPORT", "3306")
	dbinstance := getEnv("MPDBINSTANCE", "mplinksters")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbinstance)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func DBConnection() (*sql.DB, error) {
	dblocation := getEnv("MPDB", "AWS")
	var db *sql.DB
	var err error

	log.Info().Msg("Connecting to DB...")

	if dblocation == "AWS" {
		log.Info().Msg("Using AWS DB...")
		db, err = sql.Open("mysql", getDSNAWS())
		if err != nil {
			return nil, err
		}
	} else {
		log.Info().Msg("Using Local DB...")
		db, err = sql.Open("mysql", getDSNEnv())
		if err != nil {
			return nil, err
		}
	}

	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(64)
	db.SetConnMaxLifetime(time.Minute)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Connected.")

	return db, nil
}
