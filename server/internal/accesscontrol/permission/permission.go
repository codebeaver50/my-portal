// Package permission はアクセス管理で扱うリソース（feature等）と権限レベルを定義する。
// DBやHTTPに依存しない純粋な定義・計算のみを置き、service・middlewareの双方から使う。
package permission

// Level はリソースに対する権限レベル。none < read < write の順に強く、write は read を含む。
type Level int

const (
	None Level = iota
	Read
	Write
)

// ParseLevel はDB・APIでの文字列表現（"read" / "write"）を Level に変換する。
// それ以外の値（空文字を含む）は権限なしとして扱う。
func ParseLevel(value string) Level {
	switch value {
	case "read":
		return Read
	case "write":
		return Write
	default:
		return None
	}
}

// LookupLevel はAPIのリクエストで受け取った文字列（"none" / "read" / "write"）を Level に変換する。
// ParseLevel と違い、それ以外の値は ok=false にする（不正な値を権限なしとして黙って受け入れないため）。
func LookupLevel(value string) (level Level, ok bool) {
	switch value {
	case "none":
		return None, true
	case "read":
		return Read, true
	case "write":
		return Write, true
	default:
		return None, false
	}
}

// String はDB・APIでの文字列表現を返す。
func (l Level) String() string {
	switch l {
	case Read:
		return "read"
	case Write:
		return "write"
	default:
		return "none"
	}
}

// Allows は、この権限レベルで required の操作ができるかどうかを返す。
func (l Level) Allows(required Level) bool {
	return l >= required
}

// Resource は権限を設定する対象。feature-slug（Next.js側の app/demo-portal/(features)/<feature-slug> と揃える）
// または access-control（ユーザー・ロールを管理するアクセス管理画面）。
type Resource string

const (
	BookDatabase  Resource = "book-database"
	SimpleLedger  Resource = "simple-ledger"
	SimpleCMS     Resource = "simple-cms"
	FormBuilder   Resource = "form-builder"
	ShopCSS       Resource = "shop-css"
	AccessControl Resource = "access-control"
)

// ResourceDefinition はリソースごとの設定可能な権限の上限を持つ。
type ResourceDefinition struct {
	Resource Resource
	// MaxLevel はこのリソースに設定できる最大の権限。書き込み操作のないリソース（shop-css）は read まで。
	MaxLevel Level
}

// Resources は権限を設定できるリソースの一覧（権限設定画面の表示順）。
// featureを追加したらここにも追加する。
var Resources = []ResourceDefinition{
	{Resource: BookDatabase, MaxLevel: Write},
	{Resource: SimpleLedger, MaxLevel: Write},
	{Resource: SimpleCMS, MaxLevel: Write},
	{Resource: FormBuilder, MaxLevel: Write},
	{Resource: ShopCSS, MaxLevel: Read},
	{Resource: AccessControl, MaxLevel: Write},
}

// Lookup は指定のリソースの定義を返す。未知のリソースの場合は ok=false。
func Lookup(resource Resource) (definition ResourceDefinition, ok bool) {
	for _, def := range Resources {
		if def.Resource == resource {
			return def, true
		}
	}
	return ResourceDefinition{}, false
}

// Set はリソースごとの権限レベル。キーがないリソースは None。
type Set map[Resource]Level

// Level は指定のリソースへの権限レベルを返す。
func (s Set) Level(resource Resource) Level {
	return s[resource]
}

// Allows は指定のリソースに required の操作ができるかどうかを返す。
func (s Set) Allows(resource Resource, required Level) bool {
	return s.Level(resource).Allows(required)
}

// Grant は指定のリソースの権限を、今より強い場合のみ引き上げる。
// 複数のロールを持つユーザーの権限を、ロールごとの権限の和集合（最も強いもの）にするために使う。
// 未知のリソース・上限を超える権限は、DBに不正な値が入っていても上限までに丸める。
func (s Set) Grant(resource Resource, level Level) {
	def, ok := Lookup(resource)
	if !ok {
		return
	}
	if level > def.MaxLevel {
		level = def.MaxLevel
	}
	if level > s[resource] {
		s[resource] = level
	}
}
