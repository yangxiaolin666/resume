package com.resume.company.mapper;

import com.resume.model.dto.JobPostingDTO;
import com.resume.model.dto.JobPostingQueryDTO;
import com.resume.model.entity.JobPosting;
import org.apache.ibatis.annotations.Mapper;

import java.util.List;

/**
 * 招聘信息Mapper
 */
@Mapper
public interface JobPostingMapper {

    /**
     * 查询招聘信息列表（分页）
     * @param queryDTO 查询条件
     * @return 招聘信息列表
     */
    List<JobPostingDTO> selectJobPostingsByCondition(JobPostingQueryDTO queryDTO);

    /**
     * 根据ID查询招聘信息
     * @param id 招聘信息ID
     * @return 招聘信息
     */
    JobPostingDTO selectJobPostingById(Integer id);

    /**
     * 根据ID查询招聘信息（返回实体）
     * @param id 招聘信息ID
     * @return 招聘信息
     */
    JobPosting selectJobPostingByIdForEntity(Integer id);

    /**
     * 插入招聘信息
     * @param jobPosting 招聘信息
     * @return 影响行数
     */
    int insertJobPosting(JobPosting jobPosting);

    /**
     * 更新招聘信息
     * @param jobPosting 招聘信息
     * @return 影响行数
     */
    int updateJobPosting(JobPosting jobPosting);

    /**
     * 根据ID删除招聘信息
     * @param id 招聘信息ID
     * @return 影响行数
     */
    int deleteJobPostingById(Integer id);

    /**
     * 根据企业ID查询招聘信息
     * @param comId 企业ID
     * @return 招聘信息列表
     */
    List<JobPostingDTO> selectJobPostingsByComId(Integer comId);
}
