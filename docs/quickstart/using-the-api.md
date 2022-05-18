# Using the API

Currently GRMPKG is not in a state where the UI can provide all of the necessary functionality to get it up and running.
As such you will need to use the browser to authenticate and then a client like Postman or cURL to send requests to the API.

## Obtaining a token

To obtain a token you will need to go through the standard login flow, to do this, ensure that the process is running and then head to the HTTP interface in your browser.

Once there click the login button in the top right. (Images to come soon)

Then grab the cookie that's named `grm.authentication`

## Adding an SSH Key to your user

To add an SSH key to your user run the following

```
POST /authn/ssh

ssh-rsa AAA... user@laptop
```

## Creating a namespace

A namespace is the top level grouping of repositories and is what GRMPKG will use to help users find packages

To create a namespace run the following

```
POST /api/ns/<YOUR_NAMESPACE>?public=1
```

You should replace <YOUR_NAMESPACE> with the name that you want to use and set public to 1 if you would like this to be publicly exposed (currently you should do this for all namespaces as the logic hasn't been written yet for private)

## Creating a repository

The bit that you will ultimately need for running packages is a repository, this is just a git repository, but each one needs to be claimed by one (and only one) user. This ownership is likely to change moving forward, but for now this is how it wil work.

To create a repository run the following

```
POST /api/ns/<YOUR_NAMESPACE>/r/<YOUR_REPOSITORY>?public=1
```

Again, replace the variables with the names that you would like to use. You must have created a namespace before you can create a repository in it. Also at the moment logic is only working for public repositories in this system