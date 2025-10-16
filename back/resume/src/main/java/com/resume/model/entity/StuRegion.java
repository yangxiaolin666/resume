package com.resume.model.entity;

import lombok.Data;

@Data
public class StuRegion {
    private Integer id;
    private Integer stuId;
    private String province;
    private String city;
}