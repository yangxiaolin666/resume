package com.resume.company.service;

import com.github.pagehelper.PageInfo;
import com.resume.model.dto.JobPostingDTO;
import com.resume.model.dto.JobPostingQueryDTO;
import com.resume.model.entity.JobPosting;

import java.util.List;

/**
 * 招聘信息服务接口
 */
public interface JobPostingService {

    /**
     * 发布招聘信息
     * @param jobPosting 招聘信息
     * @return 发布后的招聘信息
     */
    JobPosting publishJobPosting(JobPosting jobPosting);

    /**
     * 修改招聘信息
     * @param id 招聘信息ID
     * @param jobPosting 招聘信息
     * @return 修改后的招聘信息
     */
    JobPosting updateJobPosting(Integer id, JobPosting jobPosting);

    /**
     * 删除招聘信息
     * @param id 招聘信息ID
     * @return 是否删除成功
     */
    boolean deleteJobPosting(Integer id);

    /**
     * 查询招聘信息列表（分页）
     * @param queryDTO 查询条件
     * @return 招聘信息列表
     */
    PageInfo<JobPostingDTO> getJobPostings(JobPostingQueryDTO queryDTO);

    /**
     * 查询招聘信息详情
     * @param id 招聘信息ID
     * @return 招聘信息详情
     */
    JobPostingDTO getJobPostingDetail(Integer id);

    /**
     * 查询企业的所有招聘信息
     * @param status 招聘状态（可选）
     * @param comId 企业ID
     * @return 招聘信息列表
     */
    List<JobPostingDTO> getCompanyJobPostings(String status, Integer comId);

    /**
     * 修改招聘信息状态
     * @param id 招聘信息ID
     * @param status 新的状态
     * @return 修改后的招聘信息
     */
    JobPostingDTO updateJobPostingStatus(Integer id, String status);
}
