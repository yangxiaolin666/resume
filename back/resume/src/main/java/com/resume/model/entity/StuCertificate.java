package com.resume.model.entity;

import lombok.Data;
import java.util.Date;

@Data
public class StuCertificate {
    private Long id;
    private Long stuId;
    private String certName;
    private String level;
    private String certNo;
    private Date issueDate;
    private Date createdAt;
    private Date updatedAt;
}