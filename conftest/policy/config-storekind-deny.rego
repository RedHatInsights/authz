package main
#make sure kind always holds a value to fail if key in yaml is not there or empty
default kind = "n/a"
kind := input.store.kind
allowed = "spicedb"

deny[msg] {
  kind != allowed
  msg = sprintf("Store kind %s is not production ready. Allowed: %s", [kind,allowed])
}