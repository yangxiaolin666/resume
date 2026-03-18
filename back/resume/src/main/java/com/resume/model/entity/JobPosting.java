package com.resume.model.entity;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;
import java.util.Date;

/**
 * 招聘信息实体类
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class JobPosting implements Serializable {
    private static final long serialVersionUID = 1L;

    /**
     * 招聘信息ID
     */
    private Integer id;

    /**
     * 关联企业ID
     */
    private Integer comId;

    /**
     * 招聘职位标题
     */
    private String title;

    /**
     * 招聘职位描述
     */
    private String description;

    /**
     * 外部招聘链接
     */
    private String externalLink;

    /**
     * 招聘状态：active-招聘中，inactive-已下架
     */
    private String status;

    /**
     * 创建时间
     */
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss", timezone = "GMT+8")
    private Date createdAt;

    /**
     * 更新时间
     */
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss", timezone = "GMT+8")
    private Date updatedAt;
}
