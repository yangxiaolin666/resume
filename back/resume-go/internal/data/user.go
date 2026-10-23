// 本文件负责 user 表（登录账号）的读写，以及少量需要和 student / company 关联的账号操作。
//
// 业务定位：登录鉴权（按用户名查账号 + 校验密码哈希）、注册（插入账号）、
// 修改密码、修改昵称。调用方是 internal/service 下的 auth / user / admin 服务。
//
// 表结构上的一个关键约定：user 只存「账号」信息（用户名、密码哈希、角色），
// 学生/企业的业务资料分别放在 student / company 表，通过 user_id 关联。
// 所以改昵称要落到 student.name 或 company.com_name，而不是 user 表。

package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

// FindUserByUsername 按用户名查账号，登录流程的第一步。
//
// 入参 username 来自登录请求体，是不可信输入，因此必须用 ? 占位符绑定：
// 如果写成 "... WHERE username = '" + username + "'"，攻击者传
// ' OR '1'='1 就能绕过密码校验（SQL 注入）。用 ? 之后，驱动会把值当成纯数据发送，
// 即使内容里有引号也只会被当作字符串的一部分。
//
// 返回值：找到返回 *model.User；查不到返回 (nil, nil)，由 service 层决定
// 报“用户名或密码错误”（而不是 404，避免泄露账号是否存在）。
func (d *DB) FindUserByUsername(username string) (*model.User, error) {
	row := d.QueryRow(`SELECT user_id, username, password, user_type, create_time, update_time
		FROM user WHERE username = ?`, username)
	return scanUser(row)
}

// FindUserByID 按主键查账号，通常用于 token 解析出 userID 后回查用户信息。
// user_id 是主键，等值查询会走聚簇索引，最多扫描一行。
func (d *DB) FindUserByID(userID int) (*model.User, error) {
	row := d.QueryRow(`SELECT user_id, username, password, user_type, create_time, update_time
		FROM user WHERE user_id = ?`, userID)
	return scanUser(row)
}

// scanUser 把一行查询结果映射成 model.User。
//
// 两个学习点：
//  1. Scan 的实参数量和顺序必须与 SELECT 的列一一对应，类型也要兼容，
//     否则运行时报 "expected N destination arguments" 或类型转换错误；
//  2. 时间列可能为 NULL，所以先 Scan 到 sql.NullTime，再用 timePtr 转成 *time.Time，
//     这样 NULL 在 JSON 里就是 null，而不是一个莫名其妙的零值时间 0001-01-01。
//
// err == sql.ErrNoRows 是 QueryRow 表示「没有这行」的方式（不会返回空指针），
// 这里统一翻译成 (nil, nil)，让上层用“是否存在”而不是“是不是错误”来分支。
func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var createTime, updateTime sql.NullTime
	err := row.Scan(&u.UserID, &u.Username, &u.Password, &u.UserType, &createTime, &updateTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreateTime = timePtr(createTime)
	u.UpdateTime = timePtr(updateTime)
	return &u, nil
}

// UpdateUserPassword 覆盖密码哈希。
//
// 入参 hash 必须已经是 bcrypt 之类的哈希值，明文绝不能落到数据层。
// update_time = NOW() 用数据库时间而不是 Go 的 time.Now()，避免应用服务器和数据库
// 时钟不一致导致时间线错乱（多实例部署时这一点尤其重要）。
// WHERE user_id = ? 命中主键，只改一行。
func (d *DB) UpdateUserPassword(userID int, hash string) error {
	_, err := d.Exec(`UPDATE user SET password = ?, update_time = NOW() WHERE user_id = ?`, hash, userID)
	return err
}

// UpdateDisplayName 修改用户对外展示的名字。
// 学生改 student.name，企业改 company.com_name，账号名（user.username）保持不变。
//
// 之所以按 userType 分支到两张不同表：登录账号和展示资料是分离设计的，
// 改展示名不应该影响登录凭据。
//
// 下面这段「affected == 0 时再查一次」是一个容易踩的坑：
// MySQL 的 RowsAffected 默认返回「真正发生变化」的行数，而不是「匹配到」的行数。
// 所以把昵称改成和原来一样的值时，UPDATE 成功但 affected 可能为 0。
// 若不区分「没这行」和「值没变」，就会误报 404。
// 这里的做法是：affected == 0 时再用 existsQuery 确认资料是否存在，
// 存在就直接返回 nil（视为成功），不存在才报 404。
func (d *DB) UpdateDisplayName(userID int, userType, name string) error {
	var res sql.Result
	var existsQuery string
	var err error
	switch userType {
	case "student":
		res, err = d.Exec(`UPDATE student SET name = ? WHERE user_id = ?`, name, userID)
		existsQuery = `SELECT COUNT(*) FROM student WHERE user_id = ?`
	case "company":
		res, err = d.Exec(`UPDATE company SET com_name = ? WHERE user_id = ?`, name, userID)
		existsQuery = `SELECT COUNT(*) FROM company WHERE user_id = ?`
	default:
		return model.NewAppError(403, "该用户类型不支持修改昵称")
	}
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	// affected == 0 有两种可能：资料不存在，或新名字与旧名字相同
	var count int
	if err := d.QueryRow(existsQuery, userID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return model.NewAppError(404, "用户资料不存在，无法修改昵称")
	}
	return nil
}

// UpdateStudentPasswordByStuID 对应 Java 管理员改密 SQL：
// UPDATE user u INNER JOIN student s ON u.user_id = s.user_id SET u.password = ? WHERE s.stu_id = ?
//
// 后台管理页面拿到的是学生业务 ID（student.stu_id / student.id），不是 user_id，
// 所以这里用 MySQL 的多表 UPDATE ... JOIN 直接在一条语句里完成「由学号定位账号」，
// 比先查 user_id 再更新少一次往返。
//
// 返回值 RowsAffected 让上层知道是否真的改到了账号（0 表示学号不存在）。
func (d *DB) UpdateStudentPasswordByStuID(stuID int, hash string) (int64, error) {
	res, err := d.Exec(`UPDATE user u INNER JOIN student s ON u.user_id = s.user_id
		SET u.password = ? WHERE s.stu_id = ?`, hash, stuID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// InsertUserTx 在调用方开启的事务里插入账号，返回自增主键 user_id。
//
// 注册学生/企业时是「先插 user，再插 student/company」两步写操作，
// 必须放在同一个事务里：否则第二步失败会留下一个没有任何资料的孤儿账号。
// 所以方法签名直接接收 *sql.Tx 而不是 *DB —— 事务对象由 service 层统一 Begin/Commit/Rollback，
// 数据层只负责在这条连接上执行 SQL。
//
// LastInsertId 是 MySQL 的自增主键回填，后面插 student.user_id 时要用到它。
func (d *DB) InsertUserTx(tx *sql.Tx, username, hash, userType string) (int, error) {
	res, err := tx.Exec(`INSERT INTO user (username, password, user_type, create_time, update_time)
		VALUES (?, ?, ?, NOW(), NOW())`, username, hash, userType)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// DeleteUsersTx 批量删除账号，配合后台的「批量禁用/删除」功能。
//
// IN (?,?,?) 这种「变长参数列表」的写法：
//   - 占位符个数由 placeholders 按 len(userIDs) 动态生成；
//   - 每个具体的 ID 仍然通过 toArgs 绑定，不拼进 SQL 文本；
//   - 空切片提前返回，避免生成 IN () 这种语法错误。
//
// userIDs 一般是先删 student/company 再删 user，整体仍在同一个事务里。
func (d *DB) DeleteUsersTx(tx *sql.Tx, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := `DELETE FROM user WHERE user_id IN (` + placeholders(len(userIDs)) + `)`
	_, err := tx.Exec(query, toArgs(userIDs)...)
	return err
}
