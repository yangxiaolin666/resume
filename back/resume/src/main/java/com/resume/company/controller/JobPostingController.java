package com.resume.company.controller;

import com.github.pagehelper.PageInfo;
import com.resume.admin.mapper.CompanyMapper;
import com.resume.company.service.JobPostingService;
import com.resume.model.dto.JobPostingDTO;
import com.resume.model.dto.JobPostingQueryDTO;
import com.resume.model.entity.Company;
import com.resume.model.entity.JobPosting;
import com.resume.model.result.Result;
import com.resume.utils.JwtUtil;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 招聘信息控制器
 */
@RestController
@RequestMapping("/com")
public class JobPostingController {

    @Autowired
    private JobPostingService jobPostingService;

    @Autowired
    private CompanyMapper companyMapper;

    /**
     * 发布招聘信息
     * @param jobPosting 招聘信息
     * @param request 请求对象
     * @return 发布后的招聘信息
     */
    @PostMapping("/jobPosting")
    public Result<JobPosting> publishJobPosting(@RequestBody JobPosting jobPosting,
                                                 HttpServletRequest request) {
        // 从请求中获取userId（由拦截器验证过token）
        Integer userId = (Integer) request.getAttribute("userId");
        if (userId == null) {
            return Result.error(401, "未提供有效的用户信息");
        }

        // 根据 userId 查询企业ID
        Company company = companyMapper.selectCompanyByUserId(userId);
        if (company == null) {
            return Result.error(404, "企业信息不存在（userId: " + userId + "）");
        }

        Integer comId = company.getId();

        // 验证必填字段
        if (jobPosting.getTitle() == null || jobPosting.getTitle().trim().isEmpty()) {
            return Result.error(400, "招聘标题不能为空");
        }
        System.out.println(jobPosting.getExternalLink());

        // 设置企业ID
        jobPosting.setComId(comId);

        // 发布招聘信息
        JobPosting result = jobPostingService.publishJobPosting(jobPosting);
        return Result.success(result);
    }

    /**
     * 修改招聘信息
     * @param id 招聘信息ID
     * @param jobPosting 招聘信息
     * @param request 请求对象
     * @return 修改后的招聘信息
     */
    @PutMapping("/jobPosting/{id}")
    public Result<JobPosting> updateJobPosting(@PathVariable Integer id,
                                               @RequestBody JobPosting jobPosting,
                                               HttpServletRequest request) {
        // 从请求中获取userId（由拦截器验证过token）
        Integer userId = (Integer) request.getAttribute("userId");
        if (userId == null) {
            return Result.error(401, "未提供有效的用户信息");
        }

        // 根据 userId 查询企业ID
        Company company = companyMapper.selectCompanyByUserId(userId);
        if (company == null) {
            return Result.error(404, "企业信息不存在");
        }

        Integer comId = company.getId();

        // 验证招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingService.getJobPostingDetail(id);
        if (!existingJobPosting.getComId().equals(comId)) {
            return Result.error(403, "无权限修改该招聘信息");
        }

        // 验证必填字段
        if (jobPosting.getTitle() != null && jobPosting.getTitle().trim().isEmpty()) {
            return Result.error(400, "招聘标题不能为空");
        }

        // 修改招聘信息
        JobPosting result = jobPostingService.updateJobPosting(id, jobPosting);
        result.setComId(comId);
        return Result.success(result);
    }

    /**
     * 删除招聘信息
     * @param id 招聘信息ID
     * @param request 请求对象
     * @return 删除结果
     */
    @DeleteMapping("/jobPosting/{id}")
    public Result<Map<String, Object>> deleteJobPosting(@PathVariable Integer id,
                                                         HttpServletRequest request) {
        // 从请求中获取userId（由拦截器验证过token）
        Integer userId = (Integer) request.getAttribute("userId");
        if (userId == null) {
            return Result.error(401, "未提供有效的用户信息");
        }

        // 根据 userId 查询企业ID
        Company company = companyMapper.selectCompanyByUserId(userId);
        if (company == null) {
            return Result.error(404, "企业信息不存在");
        }

        Integer comId = company.getId();

        // 验证招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingService.getJobPostingDetail(id);
        if (!existingJobPosting.getComId().equals(comId)) {
            return Result.error(403, "无权限删除该招聘信息");
        }

        // 删除招聘信息
        boolean result = jobPostingService.deleteJobPosting(id);
        Map<String, Object> data = new HashMap<>();
        data.put("success", result);
        data.put("message", result ? "招聘信息已删除" : "删除失败");
        return Result.success(data);
    }

    /**
     * 查询招聘信息列表（分页）
     * @param queryDTO 查询条件
     * @return 招聘信息列表
     */
    @GetMapping("/jobPostings")
    public Result<PageInfo<JobPostingDTO>> getJobPostings(JobPostingQueryDTO queryDTO) {
        // 设置默认分页参数
        if (queryDTO.getPageNum() == null) {
            queryDTO.setPageNum(1);
        }
        if (queryDTO.getPageSize() == null) {
            queryDTO.setPageSize(10);
        }

        // 查询招聘信息列表
        PageInfo<JobPostingDTO> result = jobPostingService.getJobPostings(queryDTO);
        return Result.success(result);
    }

    /**
     * 查询招聘信息详情
     * @param id 招聘信息ID
     * @return 招聘信息详情
     */
    @GetMapping("/jobPosting/{id}")
    public Result<JobPostingDTO> getJobPostingDetail(@PathVariable Integer id) {
        JobPostingDTO result = jobPostingService.getJobPostingDetail(id);
        return Result.success(result);
    }

    /**
     * 查询企业的所有招聘信息（无分页）
     * @param request 请求对象
     * @return 招聘信息列表
     */
    @GetMapping("/myJobPostings")
    public Result<List<JobPostingDTO>> getCompanyJobPostings(HttpServletRequest request) {
        // 从请求中获取userId（由拦截器验证过token）
        Integer userId = (Integer) request.getAttribute("userId");
        if (userId == null) {
            return Result.error(401, "未提供有效的用户信息");
        }

        // 根据 userId 查询企业ID
        Company company = companyMapper.selectCompanyByUserId(userId);
        if (company == null) {
            return Result.error(404, "企业信息不存在");
        }

        Integer comId = company.getId();

        // 获取状态筛选
        String status = request.getParameter("status");

        // 查询企业的所有招聘信息（按企业ID筛选）
        List<JobPostingDTO> result = jobPostingService.getCompanyJobPostings(status, comId);
        return Result.success(result);
    }

    /**
     * 启用/禁用招聘信息
     * @param id 招聘信息ID
     * @param status 新的状态
     * @param request 请求对象
     * @return 修改后的招聘信息
     */
    @PatchMapping("/jobPosting/{id}/status")
    public Result<JobPostingDTO> updateJobPostingStatus(@PathVariable Integer id,
                                                         @RequestBody Map<String, String> params,
                                                         HttpServletRequest request) {
        // 从请求中获取userId（由拦截器验证过token）
        Integer userId = (Integer) request.getAttribute("userId");
        if (userId == null) {
            return Result.error(401, "未提供有效的用户信息");
        }

        // 根据 userId 查询企业ID
        Company company = companyMapper.selectCompanyByUserId(userId);
        if (company == null) {
            return Result.error(404, "企业信息不存在");
        }

        Integer comId = company.getId();

        // 验证招聘信息是否存在
        JobPostingDTO existingJobPosting = jobPostingService.getJobPostingDetail(id);
        if (!existingJobPosting.getComId().equals(comId)) {
            return Result.error(403, "无权限修改该招聘信息");
        }

        // 获取新状态
        String status = params.get("status");
        if (status == null || status.isEmpty()) {
            return Result.error(400, "状态不能为空");
        }

        // 修改招聘信息状态
        JobPostingDTO result = jobPostingService.updateJobPostingStatus(id, status);
        return Result.success(result);
    }
}
