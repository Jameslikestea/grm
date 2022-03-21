package repos

test_internal_namespace_used {
    not valid_name with input as {"repo": {"name": "_internal.users"}}
}

test_git_specified {
    not valid_name with input as {"repo": {"name": "users.git"}}
}