# Authentication tokens

kube-apiserver supports several [authentication strategies][] but makes no guarantee about the order
in which they are tried. This could result in tokens intended for other authentication providers to
end up in the review queue of kubernetes-gitlab-authn. The tokens provided as part of the review process
are used as authentication input for requests against the Gitlab API when retrieving user information.
To prevent unecessary API calls kubernetes-gitlab-authn supports pre-validating to short circuit the
authentication process for obviously invalid tokens (invalid in the context of the Gitlab API).

Gitlab has [several token concepts][], each used and issued in a different context. In addition to that,
the token format can be [customized][]. To ensure the validation preflight logic does not reject
valid tokens, the list of valid tokens can be configured using the `gitlab.token_prefixes` directive.
Any token which does not have one of the given prefixes results in early authentication rejection,
i.e. no authentication information is cached nor is any request against the Gitlab API issues.

By default *Personal access tokens* (`glpat-`), *OAuth Application Secrets* (`gloas-`),
and *SCIM Tokens* (`glsoat-`) are allowed.

[authentication strategies]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies
[several token concepts]: https://docs.gitlab.com/ee/security/tokens/#token-prefixes
[customized]: https://gitlab.com/gitlab-org/gitlab/-/issues/388379
