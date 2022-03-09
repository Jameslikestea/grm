package repos

test_public_repo_anon {
    read with input as {"repo": {"public": true}}
}

test_private_repo_anon {
    not read with input as {"repo": {"public": false}}
}

test_private_repo_read_perms {
    read with input as {
        "repo": {
            "name": "galaxies/andromeda",
        },
        "uid": "1234",
        "repo_permissions": [
            {
                "uid": "1234",
                "repo_name": "galaxies/andromeda",
                "read": true,
                "write": false,
                "admin": false,
            }
        ]
    }
}

test_private_repo_read_invalid_perms {
    not read with input as {
        "repo": {
            "name": "galaxies/andromeda",
        },
        "uid": "1234",
        "repo_permissions": [
            {
                "uid": "1234",
                "repo_name": "galaxies/andromeda",
                "read": false,
                "write": false,
                "admin": false,
            }
        ]
    }
}