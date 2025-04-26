package constants

type HandlerResult int

const (
	NoHandlerResult     HandlerResult = iota + 1 // not handled
	PassHandlerResult                            // pass
	RefuseHandlerResult                          // refuse
	CancelHandlerResult                          // cancel
)

type GroupRoleLevel int

const (
	CreatorGroupRoleLevel GroupRoleLevel = iota + 1
	ManagerGroupRoleLevel
	AtLargeGroupRoleLevel
)

type GroupJoinSource int

const (
	InviteGroupJoinSource GroupJoinSource = iota + 1
	PutInGroupJoinSource
)
