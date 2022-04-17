package repos

create {
    input.create_repo.name != input.repo.name
    input.create_repo.namespace != input.repo.namespace
}