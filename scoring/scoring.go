package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mariners/player"
	"mariners/team"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Score struct {
	Player player.Player
	TeamID team.Team
	Scores [9]int
}
type MPAverage struct {
	Player  player.Player
	Average float64
	Last20  float64
	Rank    int64
	Rounds  int64
}

type MPAverages []MPAverage

type apiKey map[string]string

func getSecret() string {
	secretName := "ennead"
	region := "us-east-1"

	//Create a Secrets Manager client
	sess, err := session.NewSession()
	if err != nil {
		// Handle session creation error
		fmt.Println(err.Error())
		return ""
	}
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

	// Decrypts secret using the associated KMS key.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		return ""
	}

	k := apiKey{}
	log.Println(secretString)
	err = json.Unmarshal([]byte(secretString), &k)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return k["GoogleAPIKey"]
}

func (mp *MPAverage) GetAverage() error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithAPIKey(getSecret()))
	if err != nil {
		return err
	}

	sheetId := "1H2lhew-tk1jWQg8cDI-hMiR2xDtBDkw-862s_pP82uI"
	readRange := "2022!A2:E45"

	resp, err := srv.Spreadsheets.Values.Get(sheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			if row[1] == mp.Player.PreferredName {
				mp.Rank = row[0].(int64)
				mp.Last20 = row[2].(float64)
				mp.Average = row[3].(float64)
				mp.Rounds = row[4].(int64)
			}
		}
	}

	return nil
}

func GetAverages() (MPAverages, error) {
	as := make(MPAverages, 0)

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithAPIKey(getSecret()))
	if err != nil {
		return as, err
	}

	sheetId := "1H2lhew-tk1jWQg8cDI-hMiR2xDtBDkw-862s_pP82uI"
	readRange := "2022!A2:E45"

	resp, err := srv.Spreadsheets.Values.Get(sheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return as, err
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return as, err
	} else {
		for _, row := range resp.Values {
			var avg MPAverage
			avg.Rank, err = strconv.ParseInt(row[0].(string), 10, 64)
			if err != nil {
				return as, err
			}
			err = avg.Player.GetPlayerByPreferredName(row[1].(string))
			if err != nil {
				log.Printf("No name in scores spreadsheet matches \"%s\", or maybe something else: %s", row[1].(string), err)
			}
			avg.Average, err = strconv.ParseFloat(row[3].(string), 64)
			if err != nil {
				return as, err
			}
			avg.Last20, err = strconv.ParseFloat(row[2].(string), 64)
			if err != nil {
				return as, err
			}
			avg.Rounds, err = strconv.ParseInt(row[4].(string), 10, 64)
			if err != nil {
				return as, err
			}
			as = append(as, avg)
		}
	}

	return as, nil
}
