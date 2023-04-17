host=$1

function fail() {
    echo "$1"
    exit 1
}

function assert() {
    condition=$1
    message=$2

    if [ ! $condition ]; then
        fail "FAIL! Condition: $condition Detail: $message"
    fi
}

msg='Granting license to synthetic_test (should succeed)'
curl --fail -X POST https://$host/v1alpha/orgs/o1/licenses/smarts -H "Content-Type: application/json" -H "Authorization: token" -d '{"assign": ["synthetic_test"]}' || fail "Failed request: $msg"


msg='Getting number of seats available - should be less than the license allows'
previousAvailable=`( curl --silent --fail https://$host/v1alpha/orgs/o1/licenses/smarts -H "Authorization: token" || fail "Failed request: $msg") | jq ".seatsAvailable"`
assert "$previousAvailable -lt 10" "$msg"

echo "Waiting for quantization interval"
sleep 5

msg='Checking access for synthetic_test (should succeed)'
ret=`( curl --silent --fail -X POST https://$host/v1alpha/check -H "Content-Type: application/json" -H "Authorization: token" -d '{"subject": "synthetic_test", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = true" "$msg"

msg='Checking if synthetic_test is included in the list of assigned users'
ret=`( curl --silent --fail -H "Authorization: token" https://$host/v1alpha/orgs/o1/licenses/smarts/seats || fail "Failed request: $msg") | jq 'any(.users[]; .id == "synthetic_test")'`
assert "$ret = true" "$msg"

msg='Revoking license for synthetic_test (should succeed)'
curl --fail -X POST https://$host/v1alpha/orgs/o1/licenses/smarts -H "Content-Type: application/json" -H "Authorization: token" -d '{"unassign": ["synthetic_test"]}' || fail "Failed request: $msg"

echo "Waiting for quantization interval"
sleep 5

msg='Getting license counts again - one more should be available'
newAvailable=`( curl --silent --fail https://$host/v1alpha/orgs/o1/licenses/smarts -H "Authorization: token" || fail "Failed request: $msg" ) | jq ".seatsAvailable"`
assert "$previousAvailable -lt $newAvailable" "$msg"

msg="Checking access for synthetic_test again (should return false)"
ret=`( curl --silent --fail -X POST https://$host/v1alpha/check -H "Content-Type: application/json" -H "Authorization: token" -d '{"subject": "synthetic_test", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert "$ret = false" "$msg"

echo "PASS"
