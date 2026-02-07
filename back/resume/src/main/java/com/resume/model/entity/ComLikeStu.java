package com.resume.model.entity;


import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ComLikeStu {
    private Integer id;
    private Integer comId;
    private Integer stuId;

    private LocalDateTime createTime;
}
