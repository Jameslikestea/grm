package namespaces

default read = false
default write = false
default admin = false

read = true {
    input.namespace.public == true
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
    namespace_permissions[_].write == true
}

write = true {
    namespace_permissions[_].admin == true
}

admin = true {
    namespace_permissions[_].admin == true
}

namespace_permissions[perm] {
    perm := input.namespace_permissions[_]
    perm.uid == input.uid
    perm.namespace == input.namespace.name
}