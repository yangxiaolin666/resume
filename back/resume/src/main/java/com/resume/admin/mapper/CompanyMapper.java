package com.resume.admin.mapper;

import com.resume.model.entity.Company;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

/**
 * 企业Mapper接口
 */
@Mapper
public interface CompanyMapper {

    /**
     * 根据ID查询企业
     * @param id 企业ID
     * @return 企业信息
     */
    Company selectCompanyById(@Param("id") Integer id);

    /**
     * 根据用户ID查询企业
     * @param userId 用户ID
     * @return 企业信息
     */
    Company selectCompanyByUserId(@Param("userId") Integer userId);
}
