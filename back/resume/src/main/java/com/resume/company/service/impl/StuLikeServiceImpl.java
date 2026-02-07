package com.resume.company.service.impl;

import com.github.pagehelper.PageHelper;
import com.github.pagehelper.PageInfo;
import com.resume.company.mapper.StuLikeMapper;
import com.resume.company.service.StuLikeService;
import com.resume.model.dto.ComLikeStuDTO;
import com.resume.model.entity.ComLikeStu;
import com.resume.model.entity.Student;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.sql.Date;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
public class StuLikeServiceImpl implements StuLikeService {

    @Autowired
    private StuLikeMapper stuLikeMapper;


    @Transactional(rollbackFor = Exception.class)
    @Override
    public List<ComLikeStu> insertLikeStudent(List<ComLikeStu> list) {
        List<ComLikeStu> listHad = new ArrayList<>();
        if (list == null || list.isEmpty()) {
            return new ArrayList<>();
        }
        // 1. 校验学生ID和企业ID的有效性
        for (ComLikeStu item : list) {
            if (!stuLikeMapper.existsStudentById(item.getStuId())) {
                throw new IllegalArgumentException("学生ID不存在: " + item.getStuId());
            }
            if (!stuLikeMapper.existsCompanyById(item.getComId())) {
                throw new IllegalArgumentException("企业ID不存在: " + item.getComId());
            }
        }

        // 2. 批量查询已收藏的学生
        List<ComLikeStu> existingLikes = stuLikeMapper.searchLikeStudents(list);
        listHad.addAll(existingLikes);

        // 3. 过滤掉已收藏的学生，避免重复插入
        List<ComLikeStu> toInsert = list.stream()
                .filter(item -> existingLikes.stream()
                        .noneMatch(existing -> existing.getStuId().equals(item.getStuId()) &&
                                existing.getComId().equals(item.getComId())))
                .collect(Collectors.toList());

        // 4. 插入新收藏记录
        for (ComLikeStu item : toInsert) {
            item.setCreateTime(java.time.LocalDateTime.now());
        }

        if (!toInsert.isEmpty()) {
            int insertedRows = stuLikeMapper.insertLikeStudent(toInsert);
            if (insertedRows != toInsert.size()) {
                throw new RuntimeException("部分收藏记录插入失败");
            }
        }
//        log.info("开始插入收藏记录，参数：{}", list);
//        log.info("已存在的收藏记录：{}", listHad);
//        log.info("待插入的新记录：{}", toInsert);
        // 5. 返回已收藏的学生列表

        return listHad;
    }


    @Transactional(rollbackFor = Exception.class)
    @Override
    public Integer deleteLikeStudent(List<ComLikeStu> list) {
        List<ComLikeStu> listSearch = stuLikeMapper.searchLikeStudents(list);
        for (ComLikeStu item : list) {
            log.info("开始删除收藏记录，参数：{}", item);
        }
        for (ComLikeStu item : listSearch) {
            log.info("已存在的收藏记录：{}", item);
        }
        if (listSearch.isEmpty()) {
            return list.size();
        }
        stuLikeMapper.deleteLikeStudent(listSearch);
        //返回删除失败的学生数量
        return list.size() - listSearch.size();



    }



    @Override
    public PageInfo<ComLikeStu> searchCompaniesLike(ComLikeStuDTO comLikeStuDTO) {

        PageHelper.startPage(comLikeStuDTO.getPageNum(), comLikeStuDTO.getPageSize());

        List<ComLikeStu> stu = stuLikeMapper.searchCompaniesLike(comLikeStuDTO);

        return new PageInfo<>(stu);

    }

    @Override
    public boolean searchIfLikeStudents(ComLikeStu ComLikeStu) {

        List<ComLikeStu> list = stuLikeMapper.searchLikeStudents(List.of(ComLikeStu));
        return !list.isEmpty();
    }
}
