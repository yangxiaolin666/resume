package com.resume.model.entity;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;

/**
 * 企业实体类
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Company implements Serializable {
    private static final long serialVersionUID = 1L;

    /**
     * 企业ID
     */
    private Integer id;

    /**
     * 企业名称
     */
    private String comName;

    /**
     * 企业所在地区
     */
    private String comRegion;

    /**
     * 关联用户ID
     */
    private Integer userId;
}
