package access

const (
	// GitlabKeyNamespace is the key namespace used in a user's "extra"
	// to represent the various Gitlab specific account attributes
	GitlabKeyNamespace = "gitlab-authn.kubernetes.io/"
	// GitlabAttributesKey is the key used in a user's "extra" to specify
	// the Gitlab specific account attributes
	GitlabAttributesKey = GitlabKeyNamespace + "user-attributes"
	// GitlabGroup is the group prefix for groups based on user attributes
	GitlabGroup = "gitlab"
)

const (
	// Attribute2fa is the extra value added to authentication objects
	// when the user has 2FA enabled
	Attribute2fa = "2fa"
	// AttributeBot is the extra value added to authentication objects
	// when the user is a robot account
	AttributeBot = "bot"
	// AttributeAdmin is the extra value added to authentication objects
	// when the user is an administrator
	AttributeAdmin = "admin"
	// AttributeAuditor is the extra value added to authentication objects
	// when the user is an auditor
	AttributeAuditor = "auditor"
	// AttributeExternal is the extra value added to authentication objects
	// when the user is marked as external
	AttributeExternal = "external"
	// AttributePrivate is the extra value added to authentication objects
	// when the user account has the private flag set
	AttributePrivate = "private"
	// AttributeLocked is the extra value added to authentication objects
	// when the user account has been locked
	AttributeLocked = "locked"
	// AttributePristine is the extra value added to authentication objects
	// when the user has not yet confirmed their account
	AttributePristine = "pristine"
)

const (
	// Group2fa is the pseudo group added to authentication objects
	// when the user has 2FA enabled
	Group2fa = GitlabGroup + ":" + Attribute2fa
	// GroupBot is the pseudo group added to authentication objects
	// when the user is a robot account
	GroupBot = GitlabGroup + ":" + AttributeBot
	// GroupAdmin is the pseudo group added to authentication objects
	// when the user is an administrator
	GroupAdmin = GitlabGroup + ":" + AttributeAdmin
	// GroupAuditor is the pseudo group added to authentication objects
	// when the user is an auditor
	GroupAuditor = GitlabGroup + ":" + AttributeAuditor
	// GroupExternal is the pseudo group added to authentication objects
	// when the user is marked as external
	GroupExternal = GitlabGroup + ":" + AttributeExternal
	// GroupPrivate is the pseudo group added to authentication objects
	// when the user account has the private flag set
	GroupPrivate = GitlabGroup + ":" + AttributePrivate
	// GroupLocked is the pseudo group added to authentication objects
	// when the user account has been locked
	GroupLocked = GitlabGroup + ":" + AttributeLocked
	// GroupPristine is the pseudo group added to authentication objects
	// when the user has not yet confirmed their account
	GroupPristine = GitlabGroup + ":" + AttributePristine
)
