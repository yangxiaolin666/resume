package com.resume.company.service.impl;

import com.github.pagehelper.PageHelper;
import com.github.pagehelper.PageInfo;
import com.resume.admin.mapper.CompanyMapper;
import com.resume.company.mapper.JobPostingMapper;
import com.resume.company.service.JobPostingService;
import com.resume.model.dto.JobPostingDTO;
import com.resume.model.dto.JobPostingQueryDTO;
import com.resume.model.entity.Company;
import com.resume.model.entity.JobPosting;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

/**
 * 招聘信息服务实现类
 */
@Service
public class JobPostingServiceImpl implements JobPostingService {

    @Autowired
    private JobPostingMapper jobPostingMapper;

    @Autowired
    private CompanyMapper companyMapper;

    @Override
    @Transactional
    public JobPosting publishJobPosting(JobPosting jobPosting) {
        // 验证必填字段
        if (jobPosting.getTitle() == null || jobPosting.getTitle().trim().isEmpty()) {
            throw new RuntimeException("招聘标题不能为空");
        }
        //System.out.println(jobPosting.getExternalLink());

        // 验证企业ID是否存在
        if (jobPosting.getComId() == null) {
            throw new RuntimeException("企业ID不能为空");
        }
        Company company = companyMapper.selectCompanyById(jobPosting.getComId());
        if (company == null) {
            throw new RuntimeException("企业ID无效，企业不存在");
        }

        // 设置默认状态
        if (jobPosting.getStatus() == null) {
            jobPosting.setStatus("active");
        }

        // 插入招聘信息
        jobPostingMapper.insertJobPosting(jobPosting);

        // 重新查询数据库，获取完整的招聘信息（包括 createdAt、updatedAt 等字段）
        JobPosting createdJobPosting = jobPostingMapper.selectJobPostingByIdForEntity(jobPosting.getId());
        return createdJobPosting;
    }

    @Override
    @Transactional
    public JobPosting updateJobPosting(Integer id, JobPosting jobPosting) {
        // 查询招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingMapper.selectJobPostingById(id);
        if (existingJobPosting == null) {
            throw new RuntimeException("招聘信息不存在");
        }

        // 验证必填字段
        if (jobPosting.getTitle() != null && jobPosting.getTitle().trim().isEmpty()) {
            throw new RuntimeException("招聘标题不能为空");
        }

        // 更新招聘信息
        jobPosting.setId(id);
        jobPostingMapper.updateJobPosting(jobPosting);

        // 重新查询数据库，获取更新后的完整招聘信息（包括 updatedAt 字段）
        JobPosting updatedJobPosting = jobPostingMapper.selectJobPostingByIdForEntity(id);
        return updatedJobPosting;
    }

    @Override
    @Transactional
    public boolean deleteJobPosting(Integer id) {
        // 验证招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingMapper.selectJobPostingById(id);
        if (existingJobPosting == null) {
            throw new RuntimeException("招聘信息不存在");
        }

        // 删除招聘信息
        int result = jobPostingMapper.deleteJobPostingById(id);
        return result > 0;
    }

    @Override
    public PageInfo<JobPostingDTO> getJobPostings(JobPostingQueryDTO queryDTO) {
        // 设置分页参数
        PageHelper.startPage(queryDTO.getPageNum(), queryDTO.getPageSize());

        // 执行查询
        List<JobPostingDTO> jobPostings = jobPostingMapper.selectJobPostingsByCondition(queryDTO);

        // 返回分页结果
        return new PageInfo<>(jobPostings);
    }

    @Override
    public JobPostingDTO getJobPostingDetail(Integer id) {
        // 查询招聘信息详情
        JobPostingDTO jobPosting = jobPostingMapper.selectJobPostingById(id);

        if (jobPosting == null) {
            throw new RuntimeException("招聘信息不存在");
        }

        return jobPosting;
    }

    @Override
    public List<JobPostingDTO> getCompanyJobPostings(String status, Integer comId) {
        // 查询企业的所有招聘信息（按企业ID筛选）
        JobPostingQueryDTO queryDTO = new JobPostingQueryDTO();
        queryDTO.setStatus(status);
        queryDTO.setPageNum(1);
        queryDTO.setPageSize(Integer.MAX_VALUE);
        queryDTO.setComId(comId);

        List<JobPostingDTO> result = jobPostingMapper.selectJobPostingsByCondition(queryDTO);
        return result;
    }

    @Override
    @Transactional
    public JobPostingDTO updateJobPostingStatus(Integer id, String status) {
        // 验证状态值
        if (!"active".equals(status) && !"inactive".equals(status)) {
            throw new RuntimeException("无效的状态值，只能是active或inactive");
        }

        // 查询招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingMapper.selectJobPostingById(id);
        if (existingJobPosting == null) {
            throw new RuntimeException("招聘信息不存在");
        }

        // 更新状态
        JobPosting jobPosting = new JobPosting();
        jobPosting.setId(id);
        jobPosting.setStatus(status);
        jobPostingMapper.updateJobPosting(jobPosting);

        // 返回更新后的招聘信息
        return jobPostingMapper.selectJobPostingById(id);
    }
}
