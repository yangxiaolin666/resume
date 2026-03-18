package com.resume.model.dto;

import lombok.Data;

import java.io.Serializable;

/**
 * 招聘信息查询DTO
 */
@Data
public class JobPostingQueryDTO implements Serializable {
    private static final long serialVersionUID = 1L;

    /**
     * 招聘状态：active-招聘中，inactive-已下架
     */
    private String status;

    /**
     * 当前页码
     */
    private Integer pageNum;

    /**
     * 每页大小
     */
    private Integer pageSize;

    /**
     * 关键词搜索（职位标题模糊查询）
     */
    private String keyword;
}
