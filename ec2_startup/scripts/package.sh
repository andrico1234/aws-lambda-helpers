#!/bin/bash
FILE_NAME=ec2_startup

echo "Building Lambda..."
GOOS=linux go build -o ../handler/$FILE_NAME ../handler/$FILE_NAME.go

echo "Packaging Lambda..."
zip ../handler/handler.zip ../handler/$FILE_NAME

echo "Deploying Lambda..."
aws lambda create-function \
  --function-name $FILE_NAME \
  --runtime go1.x \
  --memory 128 \
  --role arn:aws:iam::852207075430:role/lambda_basic_execution \
  --zip-file fileb://../handler/handler.zip \
  --handler handler/$FILE_NAME

echo "Done!"