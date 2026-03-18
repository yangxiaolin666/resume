package com.resume.admin.service.impl;

import com.resume.admin.mapper.SkillsMapper;
import com.resume.admin.service.SkillsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.sql.Date;
import java.util.Set;

@Service
public class SkillsServiceImpl implements SkillsService {

    @Autowired
    private SkillsMapper skillsMapper;
    public int addSkillAll(String[] skill) {

        Date Now = new Date(System.currentTimeMillis());
        int successNum = skillsMapper.addSkills(skill, Now);
        return successNum;
    }

    @Override
    public String getAllSkill() {

        String skills = skillsMapper.getAllSkill();
        Set<String> skillSet = Set.of(skills.split(","));
        StringBuilder result = new StringBuilder();
        for (String skill : skillSet) {
            result.append(skill).append(",");
        }
        return result.substring(0, result.length() - 1);

    }

    @Override
    public int deleteSkill(String[] skill) {
        int deleteNum =skillsMapper.deleteSkill(skill);
        return deleteNum;
    }
}
