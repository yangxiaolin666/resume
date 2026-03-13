package com.resume.company.controller;


import com.resume.company.service.FavoriteService;
import com.resume.company.service.StuLikeService;
import com.resume.model.dto.ComLikeStuDTO;
import com.resume.model.entity.ComLikeStu;
import com.resume.model.result.Result;
import com.resume.utils.JwtUtil;
import jakarta.servlet.http.HttpServletRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/com")
public class LikeController {

    @Autowired
    private StuLikeService stuLikeService;

    @Autowired
    private FavoriteService favoriteService;
    @Autowired
    private RedisTemplate<Object, Object> redisTemplate;


    @GetMapping("/QueryComIdByUserId")
    public Result queryComId(HttpServletRequest request) {
        String authorizationHeader = request.getHeader("Authorization");

        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String token = authorizationHeader.substring(7); // 去掉 "Bearer " 前缀
            Integer userId = JwtUtil.getUserId(token);
            return Result.success(stuLikeService.queryComId(userId));
        }
        return Result.error(500, "token解析失败");
    }

    @PostMapping("/LikeAdd")
    public Result addLike(HttpServletRequest request, @RequestBody List<ComLikeStu> list ) {
        String authorizationHeader = request.getHeader("Authorization");
        Integer userId = null;
        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String token = authorizationHeader.substring(7); // 去掉 "Bearer " 前缀
            userId = JwtUtil.getUserId(token);
        }

//        for (int i = 0; i < list.size(); i++) {
//            log.info("添加收藏信息：{}", list.get(i));
//        }
        Integer failCount = stuLikeService.insertLikeStudent(list).size();

        //redis 缓存
        for (ComLikeStu item : list) {
            favoriteService.addFavorite(userId, item.getStuId());
        }
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
    public Result deleteLike(HttpServletRequest request, @RequestBody List<ComLikeStu> list) {

        String authorizationHeader = request.getHeader("Authorization");
        Integer userId = null;
        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String token = authorizationHeader.substring(7); // 去掉 "Bearer " 前缀
            userId = JwtUtil.getUserId(token);
        }


        Integer failCount = stuLikeService.deleteLikeStudent(list);
        redisTemplate.opsForSet().remove("user:favorite:"+userId, list);
        if (failCount == 0) {
            return Result.success();
        } else {
            return Result.error(500, "删除"+ failCount+"个学生失败，原因：学生不存在");
        }
    }
    @GetMapping("/LikeIf")
    public Result queryLikeIf(@RequestParam("com_id") Integer comId, @RequestParam("stu_id") Integer stu_id) {

        //redis 缓存
        //利用redis的set数据结构，存储学生id和收藏企业id

        ComLikeStu comLikeStu = new ComLikeStu();
        comLikeStu.setComId(comId);
        comLikeStu.setStuId(stu_id);


        return Result.success(stuLikeService.searchIfLikeStudents(comLikeStu));


    }
    @GetMapping("/LikeIfByRedis")
    public Result queryLikeIfRedis(HttpServletRequest request, @RequestParam("stuIdList") String stuIdList) {
        String authorizationHeader = request.getHeader("Authorization");

        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String token = authorizationHeader.substring(7); // 去掉 "Bearer " 前缀
            Integer userId = JwtUtil.getUserId(token);
            List<String> StuIdList = List.of(stuIdList.split(","));
            List<Integer> StuIdListInteger = new ArrayList<>();
            for (String stuId : StuIdList) {
                StuIdListInteger.add(Integer.parseInt(stuId));
            }
            log.info("userid：{}", userId);
            Map<Integer, Boolean> map = favoriteService.batchIsFavorite(userId, StuIdListInteger);
            //数据不同步？
            return Result.success(map);
        }else {
            return Result.error(500, "token解析失败");
        }

    }
}
