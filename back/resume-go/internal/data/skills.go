// 本文件负责 admin_all_skills 表：管理员维护的「技能字典」。
//
// 业务定位：学生填写简历时，技能输入框的候选项来自这里；
// 管理员在后台可以批量新增/删除技能词条。调用方是 internal/service 下的 skills / admin 服务。
//
// 注意该表在 schema.sql 里没有唯一索引（只有自增主键），
// 所以同样的技能名可能被插入多次，读取时在 Go 里做了去重（见 GetSkills）。

package data

import (
	"database/sql"
)

// AddSkills 批量新增技能，返回成功插入的条数。
//
// 实现是逐条 INSERT，没有包事务：
// 中途某条失败会直接 return，已经插入的不会回滚（返回的 count 反映实际插入数）。
// 这与学生简历编辑不同——技能字典是低频后台操作，允许部分成功，
// 失败后管理员重试即可，没必要为此引入事务开销。
// 每条都是参数化插入，技能名即使包含引号也不会破坏 SQL 结构。
func (d *DB) AddSkills(names []string) (int, error) {
	count := 0
	for _, name := range names {
		if _, err := d.Exec(`INSERT INTO admin_all_skills (skillName, create_time) VALUES (?, CURDATE())`, name); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// GetSkills 读取全部技能，供前端下拉候选。
//
// 这里在应用层做了两件事，都是因为表上没有唯一约束：
//   - 用 map 去重：历史数据里可能存在重复词条，直接返回会让下拉框出现多个同名选项；
//   - 跳过空字符串：NULL 或空技能名没有展示价值。
//
// 为什么不在 SQL 里写 SELECT DISTINCT：
//
//	DISTINCT 只能去掉“完全相同”的行，而 NULL 与空串、以及首尾带空格的脏数据
//	仍然会漏过；放在 Go 里处理也避免了数据库额外做去重排序。
//
// 更彻底的方案是给 skillName 加唯一索引 + 写入时 INSERT IGNORE，从源头杜绝重复。
//
// ORDER BY id 保持管理员录入顺序，前端展示不会每次刷新都变。
func (d *DB) GetSkills() ([]string, error) {
	rows, err := d.Query(`SELECT skillName FROM admin_all_skills ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := make(map[string]struct{})
	result := []string{}
	for rows.Next() {
		var name sql.NullString
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		value := str(name)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, rows.Err()
}

// DeleteSkills 按名称批量删除技能，返回实际删除的行数。
//
// IN (?,?,?) 的占位符个数按 names 长度动态生成，值通过 toArgsStr 绑定，
// 这样既能支持任意数量的删除目标，又不会把用户输入拼进 SQL。
// 空切片直接返回 0：否则会生成 "IN ()" 这种语法错误。
//
// 注意：表上没有唯一索引，如果同一个名称存在多条记录，这里会把它们全部删掉，
// 所以 RowsAffected 可能大于 len(names)。
func (d *DB) DeleteSkills(names []string) (int64, error) {
	if len(names) == 0 {
		return 0, nil
	}
	res, err := d.Exec(`DELETE FROM admin_all_skills WHERE skillName IN (`+placeholders(len(names))+`)`, toArgsStr(names)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// toArgsStr 把 []string 转成 []any，原因同 db.go 里的 toArgs：
// Exec 的可变参数是 ...any，[]string 不能直接展开，需要逐项装箱。
func toArgsStr(values []string) []any {
	args := make([]any, len(values))
	for i, v := range values {
		args[i] = v
	}
	return args
}
