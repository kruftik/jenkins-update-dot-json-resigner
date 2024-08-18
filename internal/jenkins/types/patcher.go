package types

type Patcher interface {
	Patch(insecureJSON *InsecureUpdateJSON) error
}
