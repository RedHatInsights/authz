baseUri=$1
idpDiscoveryUri=$2
token=""
client_id=cloud-services
scopes=api.iam.access
port=8000


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
echo TOKEN: $token
exit
