package main

empty(value) {
	count(value) == 0
}

no_violations {
	empty(deny)
}

test_denied_if_no_storekind_key_present{
	deny["Store kind n/a is not production ready. Allowed: spicedb"] with input as {"foo": "bar", "baz": {"foobar": "bazdaz"}}
}

test_denied_if_storekind_val_is_null{
	deny["Store kind null is not production ready. Allowed: spicedb"] with input as {"store": {"kind":null}}
}

test_denied_if_storekind_val_is_emptystring{
	deny["Store kind  is not production ready. Allowed: spicedb"] with input as {"store": {"kind":""}}
}

test_denied_if_storekind_val_is_multiemptystring{
	deny["Store kind     is not production ready. Allowed: spicedb"] with input as {"store": {"kind":"   "}}
}

test_denied_if_storekind_val_is_non_spicedb{
	deny["Store kind stub is not production ready. Allowed: spicedb"] with input as {"store": {"kind":"stub"}}
}

test_allow_if_storekind_val_is_spicedb{
	no_violations with input as {"store": {"kind":"spicedb"}}
}