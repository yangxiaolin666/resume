package com.resume.admin.controller;

import com.resume.admin.service.SkillsService;
import com.resume.model.result.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

@RequestMapping("/admin")
@RestController
public class SkillsController {

    @Autowired
    private SkillsService skillsService;

    /**
     * 添加学生的可以选择的技能
     */
    @Operation(summary = "添加学生可以选择的技能 (‘,’做分割的字符串)")
    @PostMapping("/addSkillAll")
    public Result addSkillAll(@RequestBody String skills) {
        String[] skill = skills.split(",");
        int successNum = skillsService.addSkillAll(skill);
        return Result.success("添加成功"+successNum+"个学生技能信息");
    }

    @Operation(summary = "获取所有学生可选技能")
    @GetMapping("/getAllSkill")
    public Result getAllSkill() {
        return Result.success(skillsService.getAllSkill());
    }

    @Operation(summary = "获取所有学生可选技能")
    @DeleteMapping("/deleteSkill")
    public Result deleteSkill(@RequestParam String skills) {
        String[] skill = skills.split(",");
        int deleteNum = skillsService.deleteSkill(skill);
        return Result.success("删除"+deleteNum+"个技能选项成功");
    }
}
