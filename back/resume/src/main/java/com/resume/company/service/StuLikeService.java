package com.resume.company.service;

import com.github.pagehelper.PageInfo;
import com.resume.model.dto.ComLikeStuDTO;
import com.resume.model.entity.ComLikeStu;
import com.resume.model.entity.StuFavorite;

import java.util.List;

public interface StuLikeService {


    /**
     * 添加收藏信息
     * @param comLikeStu 收藏信息
     * @return 返回未被收藏的学生信息
     */
    List insertLikeStudent(List<ComLikeStu> comLikeStu);


    /**
     * 删除收藏信息
     * @param comLikeStu 删除信息
     * @return 删除失败的数量
     */
    Integer deleteLikeStudent(List<ComLikeStu> comLikeStu);

    /**
     * 查找某个企业收藏的学生
     *
     * @param comLikeStuDTO 查询信息
     * @return 收藏学生列表
     */
    PageInfo<StuFavorite> searchCompaniesLike(ComLikeStuDTO comLikeStuDTO);

    /**
     * 查找企业是否收藏学生
     * @param ComLikeStu 搜索条件
     * @return
     */
    boolean searchIfLikeStudents(ComLikeStu ComLikeStu);

    Integer queryComId(Integer userId);
}
