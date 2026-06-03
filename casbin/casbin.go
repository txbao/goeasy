package casbin

import (
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"

	"github.com/txbao/goeasy/config"
)

// Enforcer RBAC/ABAC 引擎封装（不含用户/角色业务模型）。
type Enforcer struct {
	inner *casbin.Enforcer
}

func New(cfg config.CasbinCfg) (*Enforcer, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	m, err := model.NewModelFromString(defaultModel)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, err
	}
	return &Enforcer{inner: e}, nil
}

func (e *Enforcer) Check(sub, obj, act string) (bool, error) {
	if e == nil || e.inner == nil {
		return false, errors.New("casbin disabled")
	}
	return e.inner.Enforce(sub, obj, act)
}

const defaultModel = `
[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[role_definition]
g = _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
