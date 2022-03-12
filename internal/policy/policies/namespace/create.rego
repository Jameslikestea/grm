package namespaces

default create = false

create = true {
    input.create_namespace.name != input.namespace.name
}