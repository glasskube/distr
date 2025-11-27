package buildconfig

const (
	CommunityEdition  = "community"
	EnterpriseEdition = "enterprise"
)

var edition = EnterpriseEdition

func IsCommunityEdition() bool {
	return edition == CommunityEdition
}

func IsEnterpriseEdition() bool {
	return edition == EnterpriseEdition
}
