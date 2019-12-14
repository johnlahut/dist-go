# installing rabbitmq
sudo yum install erlang --enablerepo=epel
wget http://www.rabbitmq.com/releases/rabbitmq-server/v3.1.1/rabbitmq-server-3.1.1-1.noarch.rpm
sudo rpm -Uvh rabbitmq-server-3.1.1-1.noarch.rpm
sudo rabbitmq-plugins enable rabbitmq_management

# start mq and schedule it to start on node start
sudo chkconfig rabbitmq-server on
sudo service rabbitmq-server start

# be sure to create a security group with these ports opened:
# https://www.rabbitmq.com/install-rpm.html
# console http://18.206.140.49:15672/