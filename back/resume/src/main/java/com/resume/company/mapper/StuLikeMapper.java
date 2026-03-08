package com.resume.company.mapper;

import com.resume.model.dto.ComLikeStuDTO;
import com.resume.model.entity.ComLikeStu;
import com.resume.model.entity.StuFavorite;
import org.apache.ibatis.annotations.Mapper;

import java.util.List;

@Mapper
public interface StuLikeMapper {

    Integer insertLikeStudent(List<ComLikeStu> list);

    boolean existsStudentById(Integer id);

    boolean existsCompanyById(Integer id);

    //查找某个企业是否已经收藏过某个学生 返回已经收藏的学生列表
    List<ComLikeStu> searchLikeStudents(List<ComLikeStu> list);

    Integer deleteLikeStudent(List<ComLikeStu> list);

    //查找企业的收藏列表
    List<StuFavorite> searchCompaniesLike(ComLikeStuDTO comLikeStuDTO);

    Integer queryComId(Integer userId);
}
