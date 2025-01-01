# access control lists

apart from authentication, kubernetes-gitlab-authn also performs authorization
based on a simple ruleset. rules can be defined for one or more "realms".

# realms

the concept of realms allows multiple kube-apiserver instances to use the same
kubernetes-gitlab-authn service with different authorization rules. the auth
realm is selected by including it in the authentication URL endpoint
(e.g. https://gitlab-authn.example.com/authenticate/example)

An empty realm (i.e https://gitlab-authn.example.com/authenticate) is also
referred to as the *default* realm (even though there is no fallback or anything
as kube-apiserver must be given the exact endpoint it sends its review requests to)

If a request is made against a realm that has not been configured and therefor
not exist, the service responds with a 404 error.

each realm has its own set of rules. given that the service is configured with
a file using the YAML syntax, [anchors][] can be use to share common rules across realms.

```yaml
realms:
  '': # default realm
    - &common_acls
      reject_locked: true
      reject_dormant: true
      reject_pristine: true
  enclave: # multiple authz conditions
    # require 2FA set up
    - require_2fa: true
      <<: *common_acls
    # OR members of admin group
    - require_groups: [ core:admins ]
      <<: *common_acls
  lockdown: [] # noone is getting in
```

[anchors]: https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases

# rules

if no rules are configured, the service is set up to authorize everyone,
i.e. everyone with a valid token is both authenticated and authorized.

rules are evaluated using OR, where at least ONE rule must match otherwise
the user is not authorized for the respective realm.

each rule can have several criterions which in turn are evaluated using AND.

```yaml
- require_admin: true
  reject_locked: true
- require_groups:
    - kubernetes:devops
    - releng
  reject_locked: true
- require_groups:
    - kubernetes:admins
  reject_locked: true
```

the example above has two (2) rules:

* the first rule requires the user to be marked as administrator in Gitlab
* the second rule requires the user to be part of
  both the `kubernetes:devops` AND `releng` group
* the third rule requires the user to be part of
  the `kubernetes:admins` group

all rules require the Gitlab account not to be locked.

evaluation starts at the first rule and completes with the first
successful match, without processing any remaining rules.

# criteria

rules can consist of the following criteria:

* require_2fa

  Two-Factor authentication must be enabled for the account
* reject_bots

  The account must NOT be marked as [Bot][]
* reject_locked

  [Locked][] accounts are prohibited
* reject_pristine

  Users who have NOT yet confirmed their account are prohibited.
  This flag depends on your Gitlab authentication configuration,
  as users might be automatically confirmed when authenticating
  against a trusted source (e.g. LDAP)
* reject_dormant

  Gitlab reports the time it last had contact with a user.
  kubernetes-gitlab-authn can be configured with a duration
  which dictates the time window a user has to match in order
  to be considered an "active" user. Anyone not matching this
  criterion is rejected access.
* require_users

  A list of usernames to ALLOW explicitly.

  The list is evaluated using OR
* reject_users

  A list of usernames to DENY explicitly.

  The list is evaluated using OR
* require_groups

  Users must be a member of ALL of the given groups
  to be granted access.

  The list is evaluated using AND
* reject_groups

  Members of any of those groups are rejected.

  The list is evaluated using OR

[Bot]: https://docs.gitlab.com/ee/administration/internal_users.html
[Locked]: https://docs.gitlab.com/ee/security/unlock_user.html

