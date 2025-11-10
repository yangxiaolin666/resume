package com.resume.user.mapper;

import com.resume.model.entity.User;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Options;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

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
    
    /**
     * 根据用户ID查询用户
     * @param userId 用户ID
     * @return 用户信息
     */
    @Select("SELECT user_id, username, password, user_type, create_time, update_time FROM user WHERE user_id = #{userId}")
    User findByUserId(@Param("userId") Integer userId);
    
    /**
     * 插入用户并返回自增ID
     */
    @Insert("INSERT INTO user (username, password, user_type, create_time, update_time) " +
            "VALUES (#{username}, #{password}, #{userType}, NOW(), NOW())")
    @Options(useGeneratedKeys = true, keyProperty = "userId")
    int insertUser(User user);
    
    /**
     * 批量插入用户
     */
    int batchInsertUsers(@Param("users") List<User> users);
}
