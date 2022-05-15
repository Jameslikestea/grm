# Quick Start GRMPKG

In order to run GRMPKG, you will need to either download the prebuilt, Intel/Linux binary and run this, or compile for your target.

Pre-built binaries can be found on [Github](https://github.com/Jameslikestea/grm/releases)

GRMPKG is tested against S3 compatible storage (Primarily S3 and [Min.io](https://min.io)) and requires you to use Github as an authentication provider, although this will likely change to any OAuth compliant engine in the future.

## Setting up your Identity Provider

### Github

To setup Github as your authentication provider please start by following [this guide](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app) to setup an OAuth application and make a note of the Client Secret and Client ID.

## Setting up your S3 Storage

In order to use S3 as your storage option for GRM you will need an `Access Key ID` and a `Secret Access Key` these will form the two components for AWS Signed Auth.

The permissions that you will need for AWS are as follows across the entire bucket that you are using:

```text
s3:PutObject
s3:GetObject
```

You will need to create an access key and secret key with these permissions and take note of the keys.

## Setting up an RSA Key

You will also need to generate a host_key for the box, which should be shared across all instances that you want to load balance against.

You can generate this with

`openssl genrsa -out private_key.pem 2048`

I would recommend using at least a 2048 bit RSA key, however I would personally use a 4096-bit key for additional security.

## Setting up the application

To setup the application we will need the items we spoke about before as these will form the storage and identity trust relationships with the other components in the application. In order to setup the application we will need to form a yaml file `/etc/grmpkg/grmpkg.yml` which is where the application will search for the file.

The application will search in order

* The `/etc/grmpkg/` directory
* The current working directory

```yaml
authentication:
  github:
    baseurl: <YOUR HOSTNAME>
    clientid: <GITHUB CLIENT ID>
    clientsecret: <GITHUB CLIENT SECRET>
  provider: github
domain: <YOUR BASE DOMAIN>
http:
  interface: 0.0.0.0
  port: "8080"
log:
  file: true
  level: INFO
  path: /var/log/grmpkg.log
ssh:
  interface: 0.0.0.0
  keypath: <PATH TO RSA KEY>
  port: "2222"
  username: git
storage:
  s3:
    access_key: <YOUR AWS ACCESS KEY ID>
    bucket: <YOUR S3 BUCKET>
    concurrency: 5000
    endpoint: s3.<AWS REGION>.amazonaws.com
    region: <AWS REGION>
    secret_key: <YOUR AWS SECRET ACCESS KEY>
    ssl: true
  type: s3
```