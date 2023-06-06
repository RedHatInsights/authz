baseUri=$1
idpDiscoveryUri=$2
token="foo"
client_id=cloud-services
scopes=openid
port=8000

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
    #always unassign u15 on exit to avoid pollution on the next run. Also ignore all output- this will fail except when the test is exiting early before the second quantization interval
    if [ $testUserIsAssigned -eq 1 ]
    then
        echo "Exiting while u15 user should be assigned. Unassigning.."
        curl -X POST $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["u15"]}'
    fi
}

trap cleanup EXIT #always run cleanup

echo 'Setup: try to unassign u15 for a deterministic test run'
curl --fail -X POST $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["u15"]}' && echo "u15 unassigned" || echo "U15 not assigned yet. continuing..."

msg='Granting license to u15 (should succeed)'
echo $msg
curl --fail -X POST $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"assign": ["u15"]}' || fail "Failed request: $msg"

testUserIsAssigned=1 #from this point, the test user is assigned and must be unassigned on exit

msg='Getting number of seats available - should be less than the license allows'
echo $msg
previousAvailable=`( curl --silent --fail $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" || fail "Failed request: $msg") | jq ".seatsAvailable"`
assert "$previousAvailable -lt 10" "$msg"

echo "Waiting for quantization interval"
sleep 5

msg='Checking access for u15 (should succeed)'
echo $msg
ret=`( curl --silent --fail -X POST $baseUri/v1alpha/check -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"subject": "u15", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = true" "$msg"

msg='Checking if u15 is included in the list of assigned users'
echo $msg
ret=`( curl --silent --fail -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" $baseUri/v1alpha/orgs/o1/licenses/smarts/seats || fail "Failed request: $msg") | jq 'any(.users[]; .id == "u15")'`
assert "$ret = true" "$msg"

msg='Revoking license for u15 (should succeed)'
echo $msg
curl --fail -X POST $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"unassign": ["u15"]}' || fail "Failed request: $msg"

testUserIsAssigned=0 #from this point, the test user is NOT assigned, and does not need to be unassigned on exit

echo "Waiting for quantization interval"
sleep 5

msg='Getting license counts again - one more should be available'
echo $msg
newAvailable=`( curl --silent --fail $baseUri/v1alpha/orgs/o1/licenses/smarts -H "Origin: http://smoketest.test" -H "Authorization:Bearer $token" || fail "Failed request: $msg" ) | jq ".seatsAvailable"`
assert "$previousAvailable -lt $newAvailable" "$msg"

msg="Checking access for u15 again (should return false)"
echo $msg
ret=`( curl --silent --fail -X POST $baseUri/v1alpha/check -H "Origin: http://smoketest.test" -H "Content-Type: application/json" -H "Authorization:Bearer $token" -d '{"subject": "u15", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = false" "$msg"

echo "PASS"
