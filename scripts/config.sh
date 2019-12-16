# point gopath to the project
export GOPATH=$GOPATH:/Users/jlahut/UAlbany/parallel-computing/final-project/project

# ssh into the rabbitmq node
ssh -i ~/.ssh/mykey.rsa ec2-user@ec2-18-206-140-49.compute-1.amazonaws.com

# ssh into the server node
ssh -i ~/.ssh/mykey.rsa ec2-user@ec2-54-157-177-126.compute-1.amazonaws.com

# ssh into a consumer node
