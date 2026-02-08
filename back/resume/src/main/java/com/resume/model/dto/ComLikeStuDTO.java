package com.resume.model.dto;


import lombok.Data;

@Data
public class ComLikeStuDTO {

    Integer com_id;

    String comName;



    private Integer pageNum = 1;


    private Integer pageSize = 10;
}
