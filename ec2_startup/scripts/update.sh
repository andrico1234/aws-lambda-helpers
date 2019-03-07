#!/bin/bash
FILE_NAME=ec2_startup

echo "Building Lambda..."
GOOS=linux go build -o ../handler/$FILE_NAME ../handler/$FILE_NAME.go

echo "Packaging Lambda..."
rm ../handler/handler.zip

zip ../handler/handler.zip ../handler/$FILE_NAME

echo "Updating Lambda..."
aws lambda update-function-code \
  --function-name $FILE_NAME \
  --zip-file fileb://../handler/handler.zip 