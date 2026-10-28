package store

import "time"

// 预测任务的五种状态；数据库与接口都用这些字符串。
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusSucceeded = "succeeded"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// 晶体类型，与数据处理层返回的 family 一致。
const (
	FamilyIonic    = "ionic"
	FamilyCovalent = "covalent"
	FamilyMetallic = "metallic"
)

// User 是使用者。
type User struct {
	ID uint `gorm:"primaryKey"`
	// DisplayName 同时是登录名，所以要求唯一；学生自助注册时由自己填写。
	DisplayName string `gorm:"size:64;not null;uniqueIndex"`

	// PasswordHash 存 bcrypt 哈希（盐已写在哈希串里），从不存明文。
	// 空串表示还没设密码：这种账号不能登录（见 internal/auth）。
	// default:'' 让已有数据库能平滑加上这一列——老账号先置空，重置密码后再写回。
	PasswordHash string `gorm:"size:100;not null;default:''"`

	CreatedAt time.Time
}

// PredictionTask 是一次预测任务：记录输入、状态与错误信息。
type PredictionTask struct {
	ID      uint   `gorm:"primaryKey"`
	UserID  *uint  `gorm:"index"` // 可空：不登录也能预测
	Formula string `gorm:"size:64;not null"`

	// 联合索引 (status, created_at)：按状态取最早的任务时不用全表扫描。
	Status    string    `gorm:"size:16;not null;index:idx_tasks_status_created,priority:1"`
	CreatedAt time.Time `gorm:"index:idx_tasks_status_created,priority:2"`

	ErrorCode   string `gorm:"size:64"`   // 失败时的语言无关错误码
	ErrorParams string `gorm:"type:text"` // JSON：错误参数

	StartedAt  *time.Time
	FinishedAt *time.Time
	DurationMS *int64 // 计算耗时（毫秒），用于告诉学生这次算了多久

	// 外键只在这一侧声明：父子两侧都写会生成两条同名含义的约束，
	// 其中一条不带级联，删除时会被它挡住（实测会报 FOREIGN KEY constraint failed）。
	User *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// StructureResult 是任务产出的结构：晶胞参数拆成列，原子与说明存 JSON 文本。
type StructureResult struct {
	ID     uint `gorm:"primaryKey"`
	TaskID uint `gorm:"not null;uniqueIndex"` // 一个任务一行

	Family             string `gorm:"size:16;not null"`
	Prototype          string `gorm:"size:32;not null"`
	CoordinationNumber int    `gorm:"not null"`

	EdgeAAngstrom    float64 `gorm:"not null"` // 棱长 a（Å）
	EdgeBAngstrom    float64 `gorm:"not null"` // 棱长 b（Å）
	EdgeCAngstrom    float64 `gorm:"not null"` // 棱长 c（Å）
	AngleAlphaDegree float64 `gorm:"not null"` // 棱 b 与 c 的夹角 α（°）
	AngleBetaDegree  float64 `gorm:"not null"` // 棱 a 与 c 的夹角 β（°）
	AngleGammaDegree float64 `gorm:"not null"` // 棱 a 与 b 的夹角 γ（°）

	// 下面三列存 JSON 文本：结构由数据处理层定义，数据库不重复描述。
	Atoms   string `gorm:"type:text;not null"` // [{"symbol":"Na","fractional_coordinates":[0,0,0]}]
	Rule    string `gorm:"type:text;not null"` // {"code":"rule_radius_ratio","params":{…}}
	Caveats string `gorm:"type:text;not null"` // [{"code":"caveat_near_boundary","params":{…}}]

	CreatedAt time.Time

	Task *PredictionTask `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
