package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const instanceState = "running"
const tableName = "EC2Instances"

var ec2Svc = ec2.New(session.New())
var dynamodbSvc = dynamodb.New(session.New())

type ec2Instance struct {
	id      string
	state   string
	keyName string
}

func main() {
	lambda.Start(handleLambdaEvent)
}

func handleLambdaEvent(ctx context.Context) (string, error) {
	results, err := getEc2Data()

	if err != nil {
		return "", err
	}

	instances := startupInstances(results)

	if err != nil {
		return "", err
	}

	err = updateTable(instances)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Success"), nil
}

func getEc2Data() ([]map[string]*dynamodb.AttributeValue, error) {
	fmt.Println("getting ec2 data...")

	input := &dynamodb.ScanInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(instanceState),
			},
		},
		FilterExpression: aws.String("InstanceState = :s"),
		TableName:        aws.String(tableName),
	}

	result, err := dynamodbSvc.Scan(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {

			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			fmt.Println("There was an error communication with the database")
		}

		return nil, err
	}

	resultItems := result.Items

	return resultItems, nil
}

func startupInstances(instances []map[string]*dynamodb.AttributeValue) []*string {
	fmt.Println("starting up instances...")
	instanceSlice := make([]*string, 0)

	fmt.Println(instances[0]["InstanceState"].S)

	for key := range instances {
		instanceSlice = append(instanceSlice, instances[key]["Id"].S)
	}

	fmt.Println(instanceSlice)

	input := &ec2.StartInstancesInput{
		InstanceIds: instanceSlice,
	}

	ec2Svc.StartInstances(input)

	return instanceSlice
}

func updateTable(idSlice []*string) error {
	fmt.Println("updating table...")

	for _, val := range idSlice {
		input := &dynamodb.DeleteItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String(*val),
				},
			},
			TableName: aws.String(tableName),
		}

		_, err := dynamodbSvc.DeleteItem(input)

		if err != nil {
			fmt.Println(err.Error())
		}

		return err
	}

	return nil
}
