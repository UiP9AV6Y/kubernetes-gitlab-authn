# Group information

The *groups* information on a user info object are populated from the
**full path** attribute of the Gitlab Group membership information.
Slashes (`/`) are replaced by double colons (`:`) to follow the Kubernetes
naming conventions.

## Filtering

kubernetes-gitlab-authn requests group information from the Gitlab API
unfiltered by default. If a Gitlab instance happens to have a lot of
groups which are not required in the context of Kubernetes, query
[filters][list-all-groups] can be associated with outgoing requests:

| Config setting                            | Gitlab API parameter    |
|-------------------------------------------|-------------------------|
| `gitlab.group_filter.name`                | *search*                  |
| `gitlab.group_filter.owned_only`          | *owned*                   |
| `gitlab.group_filter.top_level_only`      | *top_level_only*          |
| `gitlab.group_filter.min_access_level`    | *min_access_level*        |

## Pagination

The resource used by kubernetes-gitlab-authn to fetch group information
from Gitlab is [paginated][]. The result is limited to 20 items by default
and a maximum of 100 can be fetched with a single request.

kubernetes-gitlab-authn does NOT follow the pagination system and only uses
the first batch of groups. The service does however support requesting
a different batch size using `gitlab.group_filter.limit`, which maps to the
*per_page* query parameter on the Gitlab API side. As mentioned above,
this value is limited to 100 on Gitlab side.

[list-all-groups]: https://docs.gitlab.com/ee/api/groups.html#list-all-groups
[paginated]: https://docs.gitlab.com/ee/api/rest/index.html#offset-based-pagination

