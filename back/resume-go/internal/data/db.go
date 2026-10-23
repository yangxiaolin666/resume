// Package data 是项目的数据访问层（Repository / DAO），只负责「怎么跟 MySQL 打交道」：
// 拼 SQL、绑定参数、把结果集映射成 model 结构体。它不包含业务规则，
// 业务规则（权限、参数校验、事务边界）都在 internal/service 层。
package data

import (
	"database/sql"
	"time"

	// 只导入驱动包是为了触发它的 init()，把自己注册到 database/sql 的驱动表里。
	// 因为不会直接调用这个包的任何符号，所以用 _ 作为包名（匿名导入）。
	_ "github.com/go-sql-driver/mysql"
)

// DB 组合了 *sql.DB（而不是继承），于是可以同时使用标准库的全部方法
// （Query / QueryRow / Exec / Begin …）和本层自己扩展的方法。
// 这些方法分布在 db.go、user.go、student.go 等文件里，接收者都是 *DB。
type DB struct {
	*sql.DB
}

// Open 根据 DSN 创建数据库连接池，并做一次连通性探测。
//
// 注意 sql.Open 的一个常见误区：它只做参数解析、并不会真正建立 TCP 连接，
// 所以即使 MySQL 没启动，这里也可能返回 err == nil。真正“验活”的是下面的 Ping，
// 少了这一步，错误会延迟到第一次业务查询才暴露，排查起来更麻烦。
//
// 三个连接池参数的含义（简历里也常被问到）：
//   - SetMaxOpenConns(20)：同一时刻最多 20 个连接在用，超出的请求排队等待；
//     设太大可能把 MySQL 的 max_connections 打满。
//   - SetMaxIdleConns(10)：空闲连接最多保留 10 个，避免频繁 TCP 握手 + 认证。
//   - SetConnMaxLifetime(time.Hour)：连接最长存活 1 小时就强制回收，
//     可以躲开 MySQL 的 wait_timeout 以及中间代理/LB 悄悄断开连接导致的
//     "invalid connection"。
func Open(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

// 下面四个小函数解决同一个问题：数据库里允许为 NULL 的列，Go 的 string / int / time.Time
// 都接不住（Scan 到它们会直接报错）。
//
// database/sql 为此提供了 sql.NullString / NullInt64 / NullTime：
// 它们内部多了一个 Valid 字段，用来区分「值是空字符串」和「值根本就是 NULL」——
// 这两者在业务上通常是不同含义（比如“没填性别” vs “性别填了空白”）。
// 转换规则也体现了这一点：
//   - str：NULL 视为空字符串，适合拿来展示的文本字段；
//   - strPtr / intPtr / timePtr：NULL 返回 nil 指针，适合需要区分“有没有值”的字段，
//     序列化成 JSON 时 nil 会变成 null，前端一眼就能看出来。
func str(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func intPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	// 注意这里必须取地址赋给局部变量再返回：直接 &ni.Int64 的话，
	// 拿到的是结构体内部字段的地址，语义不清晰且容易踩坑。
	v := int(ni.Int64)
	return &v
}

func strPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func timePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	v := nt.Time
	return &v
}

// placeholders 生成 n 个逗号分隔的占位符，例如 n=3 时返回 "?,?,?"。
//
// 它主要服务于 IN 批量查询：SQL 的 IN 列表长度是运行时才知道的，
// 但参数值仍然必须走占位符绑定，绝不能把值拼进 SQL 字符串里。
// 所以这里拼的只是「? 的个数」，真正的值通过 toArgs 转成 []any 交给数据库驱动，
// 驱动会走 MySQL 的预编译/转义逻辑，天然防住 SQL 注入。
func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '?')
	}
	return string(buf)
}

// toArgs 把 []int 转成 []any。
//
// 原因：Go 的方法签名里 Exec/Query 的可变参数类型是 ...any，
// 而 []int 不能直接展开成 []any（切片类型不协变），
// 必须先逐个装箱，再写成 query, args... 的形式。
// 这里只是类型转换，不改变值，也不参与 SQL 字符串拼接。
func toArgs(ids []int) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
