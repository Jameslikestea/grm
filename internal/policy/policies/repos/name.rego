package repos

default valid_name = false

valid_name = true {
    # We want to reserve the _internal namespace for the internal functions of GRM
    # this way we can create internal objects in Git.
    not internal
    # We also ensure that .git is stripped from the name, this way we can ensure
    # that the git extension is added programatically when needed.
    not git
}

internal {
    startswith(input.repo.name, "_internal")
}

git {
    endswith(input.repo.name, ".git")
}