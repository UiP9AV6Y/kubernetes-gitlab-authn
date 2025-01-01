# kubernetes-gitlab-authn deployment example

the example showcased here makes a few assumptions
about the deployment environment and serves more as a rough
guideline on how to set it up instead of being a turnkey solution.

all connections made between systems are secured with a certificate
from a public CA; provisioning custom truststores or configuring
clients to use certificates is therefor not required.

## gitlab

kubernetes-gitlab-authn performs authentication against a self-hosted
Gitlab instance under `gitlab.example.com`

users manage their [personal access tokens][] here, which are used as
authentication secret. tokens only need to have the [`read_user`][] scope
assigned. ideally users create dedicated tokens for the purpose of
authenticating against a Kubernetes cluster instead of reusing existing
tokens with potentially too many additional scopes.

[personal access tokens]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token
[`read_user`]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token

## kubernetes-gitlab-authn

kube-apiserver performs token review requests against an instance
hosted under `gitlab-authn.example.com`

personal access tokens from users sent via token review requests
from the kube-apiserver are used as authentication inputs against
the Gitlab API to retrieve information about the associated user.

```yaml
gitlab:
  address: gitlab.example.com
```

## kube-apiserver

the Kubernetes API endpoint is hosted under `kubernetes.example.com`

in order for Gitlab tokens to be accepted as authentication input,
the [webhook authentication feature][] must be enabled.

create a Kubernetes client configuration file for the kube-apiserver
to use for discovering the kubernetes-gitlab-authn webhook endpoint:

```yaml
apiVersion: v1
kind: Config

clusters:
  - name: kubernetes-gitlab-authn
    cluster:
      # include a realm in the endpoint if multiple Kubernetes clusters
      # require different authorization ACLs
      #server: https://gitlab-authn.example.com/authenticate/example
      server: https://gitlab-authn.example.com/authenticate

users:
  - name: anonymous
    user: {}

contexts:
  - name: webhook
    context:
      cluster: kubernetes-gitlab-authn
      user: anonymous

current-context: webhook
```

pass the `--authentication-token-webhook-config-file` argument to
the service startup parameters.

the `--authentication-token-webhook-cache-ttl` is optional as
kubernetes-gitlab-authn also caches authentication information. caching
it on the latter side has the advantage that users can use the same
token for authentication against multiple Kubernetes APIs without triggering
any rate limits on Gitlab side, assuming they all use the same kubernetes-gitlab-authn
instance for authentication.

providing the `--authentication-token-webhook-version` argument is also optional
as kubernetes-gitlab-authn supports any version, as it simply reuses the incoming
meta information for response payloads.

[webhook authentication feature]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication

## kubectl

clients peforming requests against the Kubernetes API must do so
with a personal access token from the Gitlab instance described above

```yaml
apiVersion: v1
kind: Config

clusters:
  - name: example
    cluster:
      server: https://kubernetes.example.com:6443/

users:
  - name: gitlab-personal-access-token
    user:
      token: glpat-SECRET-TOKEN-VALUE-0000000

contexts:
  - name: gitlab-example
    context:
      cluster: example
      user: gitlab-personal-access-token

current-context: gitlab-example
```

