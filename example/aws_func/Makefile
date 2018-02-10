build:
	GOOS=linux go build -o lambda-handler  main.go
	zip handler.zip ./lambda-handler

create:
	aws lambda create-function --region us-east-1 --function-name lambda-handler --memory 128 --role 'arn:aws:iam::$(AWS_ACCOUNT_ID):role/service-role/test' --runtime go1.x --zip-file fileb://./handler.zip --handler lambda-handler

update:
	aws lambda update-function-code --region us-east-1 --function-name lambda-handler --zip-file fileb://./handler.zip
