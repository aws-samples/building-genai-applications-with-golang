aws cloudformation create-stack \
 --stack-name cfn-ecs-network \
 --template-body file://1-network.yaml \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation update-stack \
 --stack-name cfn-ecs-network \
 --template-body file://1-network.yaml \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation create-stack \
 --stack-name cfn-ecs-load-balancer \
 --template-body file://2-load-balancer.yaml \
 --parameters '[{"ParameterKey":"NetworkStackName","ParameterValue":"cfn-ecs-network"}]' \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation update-stack \
 --stack-name cfn-ecs-load-balancer \
 --template-body file://2-load-balancer.yaml \
 --parameters '[{"ParameterKey":"NetworkStackName","ParameterValue":"cfn-ecs-network"}]' \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation create-stack \
 --stack-name cfn-ecs-cluster \
 --template-body file://3-ecs-cluster.yaml \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation create-stack \
 --stack-name cfn-ecs-task-go-bedrock \
 --template-body file://4-ecs-task.yaml \
 --capabilities CAPABILITY_NAMED_IAM

aws cloudformation create-stack \
 --stack-name cfn-go-bedrock-service \
 --template-body file://5-ecs-service.yaml \
 --parameters '[{"ParameterKey":"NetworkStackName","ParameterValue":"cfn-ecs-network"},{"ParameterKey":"ClusterStackName","ParameterValue":"cfn-ecs-cluster"},{"ParameterKey":"ALBStackName","ParameterValue":"cfn-ecs-load-balancer"},{"ParameterKey":"TaskDefinitionStackName","ParameterValue":"cfn-ecs-task-go-bedrock"},{"ParameterKey":"ContainerName","ParameterValue":"go-bedrock-app"}]' \
 --capabilities CAPABILITY_NAMED_IAM
