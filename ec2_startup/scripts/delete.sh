#!/bin/bash
functionName=ec2_startup

aws lambda delete-function --function-name $functionName