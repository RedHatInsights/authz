host=$1

function fail() {
    echo "$1"
    exit 1
}

function assert_equal() {
    expected=$1
    actual=$2
    message=$3

    if [ "$expected" != "$actual" ]; then
        fail "FAIL! Expected: $expected, got: $actual. Detail: $message"
    fi
}

msg='Granting license to synthetic_test (should succeed)'
curl --fail -X POST https://$host/v1alpha/orgs/o1/licenses/smarts -H "Content-Type: application/json" -H "Authorization: token" -d '{"assign": ["synthetic_test"]}' || fail "Failed request: $msg"

msg='Checking access for synthetic_test (should succeed)'
ret=`( curl --silent --fail -X POST https://$host/v1alpha/check -H "Content-Type: application/json" -H "Authorization: token" -d '{"subject": "synthetic_test", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert_equal "true" "$ret" "$msg"

previousAvailable=`( curl --silent --fail https://$host/v1alpha/orgs/o1/licenses/smarts -H "Authorization: token" || fail "Failed request: $msg") | jq ".seatsAvailable"`

echo "Waiting for quantization interval"
sleep 5

msg='Revoking license for synthetic_test (should succeed)'
curl --fail -X POST https://$host/v1alpha/orgs/o1/licenses/smarts -H "Content-Type: application/json" -H "Authorization: token" -d '{"unassign": ["synthetic_test"]}' || fail "Failed request: $msg"

msg='Getting license counts again - one more should be available'
newAvailable=`( curl --silent --fail https://$host/v1alpha/orgs/o1/licenses/smarts -H "Authorization: token" || fail "Failed request: $msg" ) | jq ".seatsAvailable"`
assert_equal $((previousAvailable+1)) $newAvailable "$msg"

msg="Checking access for synthetic_test again (should return false)"
ret=`( curl --silent --fail -X POST https://$host/v1alpha/check -H "Content-Type: application/json" -H "Authorization: token" -d '{"subject": "synthetic_test", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}' || fail "Failed request: $msg" ) | jq ".result"`
assert_equal 'false' $ret "$msg"

echo "PASS"