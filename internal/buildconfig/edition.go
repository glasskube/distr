package buildconfig

const (
	CommunityEdition  = "community"
	EnterpriseEdition = "enterprise"
)

var edition = EnterpriseEdition

func Edition() string {
	return edition
}

func IsCommunityEdition() bool {
	return edition == CommunityEdition
}

func IsEnterpriseEdition() bool {
	return edition == EnterpriseEdition
}
