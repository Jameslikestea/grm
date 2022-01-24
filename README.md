# GRM - Go Resource Manager

![CI](https://ci.grmpkg.com/api/v1/teams/grm/pipelines/master/badge)

This is the second attempt at building a Go Resource Manager. This project serves immutable git tags, it prevents users
from deleting projects, versions and other resources that might have become key in other projects which could cause them
to fail builds tests etc.

By using immutable tags, we can prevent this and provide more guarantees around the Go ecosystem. This is very important
for businesses and open source projects that rely on other open source dependencies. Think of GRM a bit like if NPM were
built for all languages and used Git as the protocol.