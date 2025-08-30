NETWORK="tp0_testing_net"
MESSAGE="validando-echo-server"

y sh -c
SERVER_RESPONSE=$(docker run --rn --network="$NETWORK" busybox sh -c "echo $MESSAGE | nc server 12345")

if [ "$SERVER_RESPONSE" = "$MESSAGE"]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi