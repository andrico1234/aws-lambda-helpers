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

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type instanceDetails struct {
	id      string
	keyName string
	state   string
}

func main() {
	lambda.Start(handleLambdaEvent)
}

func handleLambdaEvent(ctx context.Context) (string, error) {
	result, describeErr := describeInstances()

	if describeErr != nil {
		return "", describeErr
	}

	formattedResults := formatInstances(result)
	putItems(formattedResults)
	_, shutdownErr := shutdownInstances(formattedResults)

	if shutdownErr != nil {
		return "", shutdownErr
	}

	return fmt.Sprintf("Success"), nil
}

/*
describeInstances retrieves data for the EC2 instances whose status matches the value defined in the
instanceState constant. By default it's set to "running".

It then returns the instances as a slice.
*/
func describeInstances() ([]*ec2.Instance, error) {
	fmt.Println("retrieving instances...")

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String(instanceState),
				},
			},
		},
	}

	result, err := ec2Svc.DescribeInstances(input)

	if err != nil {
		fmt.Println("There was an error")

		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}

		return nil, err
	}

	if len(result.Reservations) == 0 {
		return nil, &errorString{"There are no instances to shutdown"}
	}

	return result.Reservations[0].Instances, nil
}

/*
formatInstances takes the results from describeInstances and maps the information to the instanceDetails struct.
*/
func formatInstances(result []*ec2.Instance) []instanceDetails {
	fmt.Println("sorting instances...")

	var instances = make([]instanceDetails, 0)

	for i := range result {
		var item = instanceDetails{
			id:      *result[i].InstanceId,
			keyName: *result[i].KeyName,
			state:   *result[i].State.Name,
		}

		instances = append(instances, item)
	}

	return instances
}

/*
updateTable takes an instanceDetails slice and updates the DynamoDB table with all of the
EC2 instances that are currently running. If the item already exists, then the function
*/
func putItems(results []instanceDetails) {
	fmt.Println("updating table...")

	for i := range results {
		input := &dynamodb.PutItemInput{
			Item: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String(results[i].id),
				},
				"KeyName": {
					S: aws.String(results[i].keyName),
				},
				"InstanceState": {
					S: aws.String(results[i].state),
				},
			},
			TableName: aws.String(tableName),
		}

		_, err := dynamodbSvc.PutItem(input)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				fmt.Println(aerr.Error())
			} else {
				fmt.Printf("There was an error adding %s details to the DynamoDB table", results[i].keyName)
			}
		} else {
			fmt.Println("There were no errors when adding data to the table")
		}
	}

	return
}

/*
	shutdownInstances takes a slice of the instanceDetails that are running and calls the shutdownAPI on them.
*/
func shutdownInstances(instances []instanceDetails) (string, error) {
	fmt.Println("shutting down instances...")

	instanceIds := make([]*string, 0)

	for i := range instances {
		instanceIds = append(instanceIds, &instances[i].id)
	}

	input := &ec2.StopInstancesInput{
		InstanceIds: instanceIds,
	}

	result, err := ec2Svc.StopInstances(input)

	if err != nil {
		fmt.Println("There was an error shutting down your EC2 Instances")
		return "", err
	}

	fmt.Println("Result: ", result)

	return "Success", nil
}
