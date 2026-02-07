package com.resume.company.controller;


import com.resume.company.service.StuLikeService;
import com.resume.model.dto.ComLikeStuDTO;
import com.resume.model.entity.ComLikeStu;
import com.resume.model.result.Result;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@Slf4j
@RestController
@RequestMapping("/com")
public class LikeController {

    @Autowired
    private StuLikeService stuLikeService;

    @PostMapping("/LikeAdd")
    public Result addLike(@RequestBody List<ComLikeStu> list ) {

//        for (int i = 0; i < list.size(); i++) {
//            log.info("添加收藏信息：{}", list.get(i));
//        }
        Integer failCount =stuLikeService.insertLikeStudent(list).size();
        if (failCount == 0) {
            return Result.success();
        }
        return Result.error(500, "添加"+ failCount+"个学生失败，原因：学生已存在");
    }

    @PostMapping("/LikeQuery")
    public Result queryLike(@RequestBody ComLikeStuDTO comDTO) {
        if (comDTO == null || comDTO.getCom_id() == null) {
            return Result.error(500, "企业ID不能为空");
        }

        return Result.success(stuLikeService.searchCompaniesLike(comDTO));
    }


    @DeleteMapping("/LikeDelete")
    public Result deleteLike(@RequestBody List<ComLikeStu> list) {

        Integer failCount = stuLikeService.deleteLikeStudent(list);
        if (failCount == 0) {
            return Result.success();
        } else {
            return Result.error(500, "删除"+ failCount+"个学生失败，原因：学生不存在");
        }
    }
    @GetMapping("/LikeQueryWithDetails")
    public Result queryLikeWithDetails(@RequestParam("com_id") Integer comId, @RequestParam("stu_id") Integer stu_id) {
        ComLikeStu comLikeStu = new ComLikeStu();
        comLikeStu.setComId(comId);
        comLikeStu.setStuId(stu_id);

        return Result.success(stuLikeService.searchIfLikeStudents(comLikeStu));


    }
}
