package com.resume.model.entity;

import lombok.Data;
import java.util.Date;

@Data
public class StuLanguage {
    private Long id;
    private Long stuId;
    private String language;
    private String level;
    private String certNo;
    private Date issueDate;
    private Date createdAt;
    private Date updatedAt;
}