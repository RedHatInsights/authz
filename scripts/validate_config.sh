#!/bin/bash
CONFIG_FILE="$1"

#validates if store is set to spicedb
validateStoreConfigIsSetToSpiceDb() {
  if [[ $(yq e '.store.kind' "$CONFIG_FILE") != "spicedb" ]]; then
    echo "Error: Store kind is not spicedb. Please use spiceDB for production usage."
    exit 1
  else
    echo "chosen store implementation spicedb is valid for production usage. Continue..."
  fi
}

# validates that at least one entry exists in the map (but "-" counts as one entry)
validateAuthConfigExists() {
  local auth_entries=$(yq e '.auth | length' "$CONFIG_FILE")
  if [[ $auth_entries -lt 1 ]]; then
    echo "Error: Auth Config must have at least one entry"
    exit 1
  else
    echo "At least one Authconfig found. Continue..."
  fi
}

# validates that at least one authconfig entry has enabled=true set
validateAuthConfigEnabled() {
  local enabled_entries=$(yq e '.auth[] | select(.enabled == true)' "$CONFIG_FILE")
  if [[ -z $enabled_entries ]]; then
    echo "Error: At least one AuthConfig entry must be set to 'enabled: true'"
    exit 1

  else
    echo "At least one AuthConfig enabled! Successfully validated config entries for production readiness. Continuing the rollout!"
  fi
}

validateStoreConfigIsSetToSpiceDb
validateAuthConfigExists
validateAuthConfigEnabled
