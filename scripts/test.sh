baseUri=$1
idpDiscoveryUri=$2
orgId=$3
userId=$4
maxSeats=$5

token=""
client_id=cloud-services
scopes=openid
port=8000

function login() {
  #Retrieve the discovery document and extract the urls we need: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse
  metadata=$(curl -s "$idpDiscoveryUri")
  authzEndpoint=$(echo "$metadata" | jq -r ".authorization_endpoint")
  tokenEndpoint=$(echo "$metadata" | jq -r ".token_endpoint")

  echo "Please open this url in a browser: $authzEndpoint?response_type=code&client_id=$client_id&redirect_uri=http://127.0.0.1:$port/&scope=$scopes" #https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest

  echo "Listening for response from identity provider.."
  #1 the response that should be sent to the browser gets piped into netcat. Note Connection: close is important because netcat won't exit until the client disconnects, and some browsers will keep a connection open for pooling purposes unless told not to.
  #2 netcat listens on the given port, sends the response to the first connection, and prints input to the pipe
  #3 awk then searches the request piped from netcat for code=yourcodehere (a querystring argument) to extract the authorization code: https://openid.net/specs/openid-connect-core-1_0.html#AuthResponse
  authzcode=$(echo -e "HTTP/1.1 200\r\nContent-Length: 26\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nPlease return to terminal." | nc -l 8000 | awk -F'[?&[:space:]]' '{for (i=1; i<=NF; i++) {split($i,p,"="); if (p[1]=="code") print p[2]}}')

  echo "Code: $authzcode"

  echo "Exchanging code for token"
  #Redeems the authorization code from the SSO for an access token by calling the token endpoint: https://openid.net/specs/openid-connect-core-1_0.html#TokenRequest
  token_response=$(curl -s -X POST -d "grant_type=authorization_code&client_id=$client_id&code=$authzcode&redirect_uri=http://127.0.0.1:$port/" $tokenEndpoint)
  token=$(echo "$token_response" | jq -r '.access_token') #Parse the access token out of the response and set it to a global: https://openid.net/specs/openid-connect-core-1_0.html#TokenResponse
}

function fail() {
    echo "$1"
    exit 1
}

function info() {
    echo "$1"
}

function assert() {
    condition=$1
    message=$2

    if [ ! $condition ]; then
        fail "FAIL! Condition: $condition Detail: $message"
    fi
}

testUserIsAssigned=0

function cleanup() {
    #always unassign $userId on exit to avoid pollution on the next run. Also ignore all output- this will fail except when the test is exiting early before the second quantization interval
    if [ $testUserIsAssigned -eq 1 ]
    then
        echo "Exiting while $userId user should be assigned. Unassigning.."
        curl -X POST $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["'$userId'"]}'
    fi
}

login
trap cleanup EXIT #always run cleanup

msg="Setup: Try unassigning $userId if assigned. should return 400 if not assigned."
echo $msg
curl --fail -X POST $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["'$userId'"]}' || info "$userId not assigned yet. continuing..."

msg="Granting license to $userId (should succeed)"
echo "Test: $msg"
curl --fail -X POST $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"assign": ["'$userId'"]}' || fail "Failed request: $msg"

testUserIsAssigned=1 #from this point, the test user is assigned and must be unassigned on exit

msg="Getting number of seats available - should be less than the license allows. License allows (param): $maxSeats"
echo "Test: $msg"
previousAvailable=`( curl --silent --fail $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" || fail "Failed request: $msg") | jq ".seatsAvailable"`
assert "$previousAvailable -lt $maxSeats" "$msg"

echo "Waiting for quantization interval"
sleep 5

msg="Checking access for $userId (should succeed)"
echo "Test: $msg"
ret=`( curl --silent --fail -X POST $baseUri/v1alpha/check -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"subject": "'$userId'", "operation": "access", "resourcetype": "license", "resourceid": "'$orgId'/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = true" "$msg"

msg="Checking if $userId is included in the list of assigned users"
echo "Test: $msg"
ret=`( curl --silent --fail -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" $baseUri/v1alpha/orgs/$orgId/licenses/smarts/seats || fail "Failed request: $msg") | jq 'any(.users[]; .id == "'$userId'")'`
assert "$ret = true" "$msg"

msg="Revoking license for $userId (should succeed)"
echo "Test: $msg"
curl --fail -X POST $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["'$userId'"]}' || fail "Failed request: $msg"

testUserIsAssigned=0 #from this point, the test user is NOT assigned, and does not need to be unassigned on exit

echo "Waiting for quantization interval"
sleep 5

msg='Getting license counts again - one more should be available'
echo "Test: $msg"
newAvailable=`( curl --silent --fail $baseUri/v1alpha/orgs/$orgId/licenses/smarts -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" || fail "Failed request: $msg" ) | jq ".seatsAvailable"`
assert "$previousAvailable -lt $newAvailable" "$msg"

msg="Checking access for $userId again (should return false)"
echo "Test: $msg"
ret=`( curl --silent --fail -X POST $baseUri/v1alpha/check -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"subject": "'$userId'", "operation": "access", "resourcetype": "license", "resourceid": "'$orgId'/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = false" "$msg"

echo "PASSED ALL TESTS"
