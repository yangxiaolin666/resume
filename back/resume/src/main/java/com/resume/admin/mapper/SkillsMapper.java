package com.resume.admin.mapper;


import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.sql.Date;

@Mapper
public interface SkillsMapper{

    /*
     * 添加学生可选技能
     */
    public int addSkills(@Param("skills") String[] skills, @Param("date") Date date);

    String getAllSkill();

    int deleteSkill(String[] skill);
}
