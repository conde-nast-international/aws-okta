# aws-okta

DEPRECATED, please use [AWS CLI + Okta Instructions](https://confluence.condenastint.com/pages/viewpage.action?pageId=64260100)

`aws-okta` allows you to authenticate with AWS using your Okta credentials.

## Installing

If you have a functional go environment, you can install with:

```bash
$ go get github.com/conde-nast-international/aws-okta
```

## Usage

### Adding Okta credentials

```bash
$ aws-okta add
```

This will prompt you for your Okta organization, username, and password.  These credentials will then be stored in your keyring for future use.   
Note: Don't append ".okta.com" to your organization name, it is automatically added by `aws-okta`.


### Exec

```bash
$ aws-okta exec <profile> -- <command>
```

Exec will assume the role specified by the given aws config profile and execute a command with the proper environment variables set.  This command is a drop-in replacement for `aws-vault exec` and accepts all of the same command line flags:

```bash
$ aws-okta help exec
exec will run the command specified with aws credentials set in the environment

Usage:
  aws-okta exec <profile> -- <command>

Flags:
  -a, --assume-role-ttl duration   Expiration time for assumed role (default 15m0s)
  -h, --help                       help for exec
  -t, --session-ttl duration       Expiration time for okta role session (default 1h0m0s)

Global Flags:
  -b, --backend string   Secret backend to use [kwallet secret-service file] (default "file")
  -d, --debug            Enable debug logging
```

### Exec for EKS and Kubernetes

`aws-okta` can also be used to authenticate `kubectl` to your AWS EKS cluster. Assuming you have [installed `kubectl`](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html), [setup your kubeconfig](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html) and [installed `aws-iam-authenticator`](https://docs.aws.amazon.com/eks/latest/userguide/configure-kubectl.html), you can now access your EKS cluster with `kubectl`. Note that on a new cluster, your Okta CLI user needs to be using the same assumed role as the one who created the cluster. Otherwise, your cluster needs to have been configured to allow your assumed role.

```bash
$ aws-okta exec <profile> -- kubectl version --short
```

Likewise, most Kubernetes projects should work, like Helm and Ark.

```bash
$ aws-okta exec <profile> -- helm version --short
```

### Configuring your aws config

`aws-okta` assumes that your base role is one that has been configured for Okta's SAML integration by your Okta admin. Okta provides a guide for setting up that integration [here](https://support.okta.com/help/servlet/fileField?retURL=%2Fhelp%2Farticles%2FKnowledge_Article%2FAmazon-Web-Services-and-Okta-Integration-Guide&entityId=ka0F0000000MeyyIAC&field=File_Attachment__Body__s).  During that configuration, your admin should be able to grab the AWS App Embed URL from the General tab of the AWS application in your Okta org.  You will need to set that value in your `~/.aws/config` file, for example:

```ini
[okta]
aws_saml_url = home/amazon_aws/0ac4qfegf372HSvKF6a3/965
```

Next, you need to set up your base Okta role.  This will be one your admin created while setting up the integration.  It should be specified like any other aws profile:

```ini
[profile okta-dev]
role_arn = arn:aws:iam::<account-id>:role/<okta-role-name>
region = <region>
```

Your setup may require additional roles to be configured if your admin has set up a more complicated role scheme like cross account roles.  For more details on the authentication process, see the internals section.

#### A more complex example

The `aws_saml_url` can be set in the "okta" ini section, or on a per profile basis. This is useful if, for example, your organization has several Okta Apps (i.e. one for dev/qa and one for prod, or one for internal use and one for integrations with third party providers). For example:

```ini
[okta]
# This is the "default" Okta App
aws_saml_url = home/amazon_aws/cuZGoka9dAIFcyG0UllG/214

[profile dev]
# This profile uses the default Okta app
role_arn = arn:aws:iam::<account-id>:role/<okta-role-name>

[profile integrations-auth]
# This is a distinct Okta App
aws_saml_url = home/amazon_aws/woezQTbGWUaLSrYDvINU/214
role_arn = arn:aws:iam::<account-id>:role/<okta-role-name>

[profile vendor]
# This profile uses the "integrations-auth" Okta app combined with secondary role assumption
source_profile = integrations-auth
role_arn = arn:aws:iam::<account-id>:role/<secondary-role-name>
```

The configuration above means that you can use multiple Okta Apps at the same time and switch between them easily.

## Backends

We use 99design's keyring package that they use in `aws-vault`.  Because of this, you can choose between different pluggable secret storage backends just like in `aws-vault`.  You can either set your backend from the command line as a flag, or set the `AWS_OKTA_BACKEND` environment variable.

For Linux / Ubuntu add the following to your bash config / zshrc etc:
```
export AWS_OKTA_BACKEND=secret-service
```

## Releasing

```bash
$ git tag v0.x.x
$ git push --tags
$ export GH_LOGIN=<github personal access code>
$ make all
$ make -f Makefile.release publish-github
```

## Analytics

`aws-okta` includes some usage analytics code which Segment uses internally for tracking usage of internal tools.  This analytics code is turned off by default, and can only be enabled via a linker flag at build time, which we do not set for public github releases.

## Internals

### Authentication process

We use the following multiple step authentication:

- Step 1 : Basic authentication against Okta
- Step 2 : MFA challenge if required
- Step 3 : Get AWS SAML assertion from Okta
- Step 4 : Assume base okta role from profile with the SAML Assertion
- Step 5 : Assume the requested AWS Role from the targeted AWS account to generate STS credentials
