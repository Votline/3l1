package contextkeys

type contextKey struct{}

var (
	UserKey = &contextKey{}
	ReqKey  = &contextKey{}
)

type UserInfo struct {
	Role   string
	UserID string
}
