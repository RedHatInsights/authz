idpDiscoveryUri=$1
ctoken_client_id=$2
client_secret=$3
client_scope=$4

token=""
port=8000



#Retrieve the discovery document and extract the urls we need: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse
metadata=$(curl -s "$idpDiscoveryUri")
tokenEndpoint=$(echo "$metadata" | jq -r ".token_endpoint")
ctoken_response=$(curl -s -X POST -d "grant_type=client_credentials&client_id=$ctoken_client_id&client_secret=$client_secret&scope=$client_scope" $tokenEndpoint)
clienttoken=$(echo "$ctoken_response" | jq -r '.access_token')
echo $clienttoken
exit

