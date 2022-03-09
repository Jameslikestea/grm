package repos

default read = false
default write = false
default admin = false

read = true {
    input.repo.public == true
}

read = true {
    repo_permissions[_].read == true
}

read = true {
    repo_permissions[_].write == true
}

read = true {
    repo_permissions[_].admin == true
}

read = true {
    namespace_permissions[_].read == true
}

read = true {
    namespace_permissions[_].write == true
}

read = true {
    namespace_permissions[_].admin == true
}

write = true {
    repo_permissions[_].write == true
}

write = true {
    repo_permissions[_].admin == true
}

write = true {
    namespace_permissions[_].write == true
}

write = true {
    namespace_permissions[_].admin == true
}

admin = true {
    repo_permissions[_].admin == true
}

admin = true {
    namespace_permissions[_].admin == true
}

repo_permissions[perm] {
    perm := input.repo_permissions[_]
    perm.uid == input.uid
    perm.repo_name == input.repo.name
}

namespace_permissions[perm] {
    perm := input.namespace_permissions[_]
    perm.uid == input.uid
    perm.namespace == input.namespace.name
}