package com.resume.user.mapper;

import com.resume.model.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

/**
 * 用户Mapper接口
 */
@Mapper
public interface UserMapper {
    
    /**
     * 根据用户名查询用户
     * @param username 用户名
     * @return 用户信息
     */
    @Select("SELECT user_id, username, password, user_type, create_time, update_time FROM user WHERE username = #{username}")
    User findByUsername(@Param("username") String username);
}
