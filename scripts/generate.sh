#!/usr/bin/env bash
## OPENAPI_FILENAME=yourapi generate.sh

# generate an API client for a service
generate_sdk() {
    local file_name=$1
    local output_path=$2
    local package_name=$3

    echo "Validating OpenAPI ${file_name}"
    npx @openapitools/openapi-generator-cli validate -i "$file_name"

    echo "Generating source code based on ${file_name}"

    # remove old generated models
    rm -Rf $OUTPUT_PATH

    npx @openapitools/openapi-generator-cli generate -g go -i \
    "$file_name" -o "$output_path" \
    --package-name="${package_name}" \
    --additional-properties=$additional_properties \
    --ignore-file-override=.openapi-generator-ignore
}

npx @openapitools/openapi-generator-cli version-manager set 5.2.0
echo "Generating SDKs"
additional_properties="generateInterfaces=true,enumClassPrefix=true"


OPENAPI_FILENAME="api/v1alpha/openapi-authz-v1_alpha.yaml"
PACKAGE_NAME="api"
OUTPUT_PATH="api/v1alpha/public"

generate_sdk $OPENAPI_FILENAME $OUTPUT_PATH $PACKAGE_NAME
